package bot

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"valarkovichMadHouseBot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot     *tgbotapi.BotAPI
	storage *storage.Storage
}

func New(bot *tgbotapi.BotAPI, storage *storage.Storage) *Handler {
	return &Handler{
		bot:     bot,
		storage: storage,
	}
}

func (h *Handler) HandleUpdate(update tgbotapi.Update) {
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
	default:
		h.reply(update.Message.Chat.ID, "Доступные команды:\n/start\n/me\n/users")

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
		"Привет, %s!\nТы зарегистрирован.\n\nКоманды:\n/start — регистрация\n/me — мои данные\n/users — список игроков",
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

	for i, user := range users {
		fullName := strings.TrimSpace(strings.Join([]string{user.FirstName, user.LastName}, " "))
		if fullName == "" {
			fullName = "Без имени"
		}

		if user.Username != "" {
			lines = append(lines, fmt.Sprintf("%d. %s (@%s)", i+1, fullName, user.Username))
		} else {
			lines = append(lines, fmt.Sprintf("%d. %s", i+1, fullName))
		}
	}

	h.reply(m.Chat.ID, strings.Join(lines, "\n"))
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
