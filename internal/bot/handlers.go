package bot

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"valarkovichMadHouseBot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot     *tgbotapi.BotAPI
	storage *storage.Storage
	adminID int64
}

func New(bot *tgbotapi.BotAPI, storage *storage.Storage, adminID int64) *Handler {
	return &Handler{
		bot:     bot,
		storage: storage,
		adminID: adminID,
	}
}

func (h *Handler) HandleUpdate(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		h.handleCallback(update.CallbackQuery)
		return
	}

	if update.Message == nil {
		return
	}

	if !update.Message.IsCommand() {
		return
	}

	switch update.Message.Command() {
	case "start":
		h.handleStart(update.Message)
	case "me":
		h.handleMe(update.Message)
	case "users":
		h.handleUsers(update.Message)
	case "delete_user":
		h.handleDeleteUser(update.Message)
	default:
		h.reply(update.Message.Chat.ID, "Доступные команды:\n/start\n/me\n/users\n/delete_user <telegram_id>")

	}

}

func (h *Handler) handleStart(m *tgbotapi.Message) {
	if m.From == nil {
		h.reply(m.Chat.ID, "Не удалось получить данные пользователя")
		return
	}

	err := h.storage.RegisterUser(
		m.From.ID,
		m.From.UserName,
		m.From.FirstName,
		m.From.LastName,
	)
	if err != nil {
		log.Printf("register user error: %v", err)
		h.reply(m.Chat.ID, "Ошибка регистрации")
		return
	}

	name := fallback(m.From.FirstName, "Валаркович")

	h.reply(m.Chat.ID, fmt.Sprintf(
		"Привет, %s!\nТы зарегистрирован.\n\nКоманды:\n/start — регистрация\n/me — мои данные\n/users — список игроков\n/delete_user <telegram_id> — удалить пользователя",
		name,
	))

}

func (h *Handler) handleMe(m *tgbotapi.Message) {
	if m.From == nil {
		h.reply(m.Chat.ID, "Не удалось получить данные пользователя")
		return
	}

	user, err := h.storage.GetUserByTelegramID(m.From.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.reply(m.Chat.ID, "Тебя еще нет в базе. Напиши /start")
			return
		}

		log.Printf("get me error: %v", err)
		h.reply(m.Chat.ID, "Ошибка чтения данных")
		return
	}

	fullName := strings.TrimSpace(strings.Join([]string{user.FirstName, user.LastName}, " "))
	if fullName == "" {
		fullName = "Без имени"
	}

	text := fmt.Sprintf(
		"Твои данные:\nID в БД: %d\nTelegram ID: %d\nИмя: %s\nUsername: %s\nДата регистрации: %s",
		user.ID,
		user.TelegramID,
		fullName,
		fallback(user.Username, "—"),
		user.RegisteredAt,
	)

	h.reply(m.Chat.ID, text)
}

func (h *Handler) handleUsers(m *tgbotapi.Message) {
	users, err := h.storage.GetAllUsers()
	if err != nil {
		log.Printf("get all users error: %v", err)
		h.reply(m.Chat.ID, "Ошибка чтения пользователей")
		return
	}

	if len(users) == 0 {
		h.reply(m.Chat.ID, "Пока нет зарегистрированных пользователей")
		return
	}

	lines := []string{"Зарегистрированные игроки:"}

	var keyboard [][]tgbotapi.InlineKeyboardButton

	for i, user := range users {
		fullName := strings.TrimSpace(strings.Join([]string{user.FirstName, user.LastName}, " "))
		if fullName == "" {
			fullName = "Без имени"
		}

		line := fmt.Sprintf("%d. %s", i+1, fullName)
		if user.Username != "" {
			line = fmt.Sprintf("%d. %s (@%s)", i+1, fullName, user.Username)
		}

		lines = append(lines, line)

		if m.From != nil && m.From.ID == h.adminID && user.TelegramID != h.adminID {
			btn := tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("Удалить %s", fullName),
				fmt.Sprintf("delete_user:%d", user.TelegramID),
			)
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(btn))
		}
	}

	msg := tgbotapi.NewMessage(m.Chat.ID, strings.Join(lines, "\n"))

	if len(keyboard) > 0 {
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	}

	_, err = h.bot.Send(msg)
	if err != nil {
		log.Printf("send users message error: %v", err)
	}
}

func (h *Handler) handleCallback(q *tgbotapi.CallbackQuery) {
	if q == nil || q.From == nil {
		return
	}

	if strings.HasPrefix(q.Data, "delete_user:") {
		h.handleDeleteUserCallback(q)
		return
	}
}

func (h *Handler) handleDeleteUserCallback(q *tgbotapi.CallbackQuery) {
	if q.From.ID != h.adminID {
		h.answerCallback(q.ID, "У тебя нет прав")
		return
	}

	idStr := strings.TrimPrefix(q.Data, "delete_user:")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.answerCallback(q.ID, "Некорректный ID")
		return
	}

	if userID == h.adminID {
		h.answerCallback(q.ID, "Нельзя удалить администратора")
		return
	}

	err = h.storage.DeleteUserByTelegramID(userID)
	if err != nil {
		log.Printf("delete user callback error: %v", err)
		h.answerCallback(q.ID, "Ошибка удаления")
		return
	}

	h.answerCallback(q.ID, "Пользователь удалён")

	if q.Message != nil {
		h.reply(q.Message.Chat.ID, fmt.Sprintf("Пользователь с telegram_id=%d удалён", userID))
	}
}

func (h *Handler) answerCallback(callbackID, text string) {
	cfg := tgbotapi.NewCallback(callbackID, text)
	_, err := h.bot.Request(cfg)
	if err != nil {
		log.Printf("answer callback error: %v", err)
	}
}

func (h *Handler) handleDeleteUser(m *tgbotapi.Message) {
	if m.From == nil {
		h.reply(m.Chat.ID, "Не удалось получить данные пользователя")
		return
	}

	if m.From.ID != h.adminID {
		h.reply(m.Chat.ID, "У тебя нет прав на удаление пользователей")
		return
	}

	args := strings.TrimSpace(m.CommandArguments())
	if args == "" {
		h.reply(m.Chat.ID, "Формат: /delete_user <telegram_id>")
		return
	}

	userID, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		h.reply(m.Chat.ID, "Неверный telegram_id")
		return
	}

	if userID == h.adminID {
		h.reply(m.Chat.ID, "Нельзя удалить администратора")
		return
	}

	err = h.storage.DeleteUserByTelegramID(userID)
	if err != nil {
		log.Printf("delete user error: %v", err)
		h.reply(m.Chat.ID, "Ошибка удаления пользователя")
		return
	}

	h.reply(m.Chat.ID, fmt.Sprintf("Пользователь с telegram_id=%d удален", userID))
}

func (h *Handler) reply(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("send message error: %v", err)
	}
}

func fallback(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}
