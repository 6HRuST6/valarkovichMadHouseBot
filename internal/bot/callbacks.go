package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
