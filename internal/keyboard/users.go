package keyboard

import (
	"fmt"
	"strings"

	"valarkovichMadHouseBot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func UsersAdminKeyboard(users []storage.User, adminID int64) *tgbotapi.InlineKeyboardMarkup {

	var rows [][]tgbotapi.InlineKeyboardButton

	for _, user := range users {

		if user.TelegramID == adminID {
			continue
		}

		name := strings.TrimSpace(strings.Join([]string{user.FirstName, user.LastName}, " "))
		if name == "" {
			name = "Без имени"
		}

		buttonText := fmt.Sprintf("❌ Удалить %s", name)
		if user.Username != "" {
			buttonText = fmt.Sprintf("❌ Удалить @%s", user.Username)
		}

		btn := tgbotapi.NewInlineKeyboardButtonData(
			buttonText,
			fmt.Sprintf("delete_user:%d", user.TelegramID),
		)

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))

	}

	if len(rows) == 0 {
		return nil
	}
	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	return &markup

}
