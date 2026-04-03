package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {

	token := os.Getenv("tg_token")

	if token == "" {
		log.Fatal("tg_bot empty")
	}

	bot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("authorized as @%s", bot.Self.UserName)

	UpdateConfig := tgbotapi.NewUpdate(0)
	UpdateConfig.Timeout = 30

	updates := bot.GetUpdatesChan(UpdateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf(
			"message from %s: %s",
			update.Message.From.UserName,
			update.Message.Text,
		)

		replyText := "Я получил сообщение:" + update.Message.Text

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, replyText)

		_, err := bot.Send(msg)
		if err != nil {
			log.Println("send message error:", err)
		}

	}
}
