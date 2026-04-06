package main

import (
	"database/sql"
	"log"
	"os"
	appbot "valarkovichMadHouseBot/internal/bot"
	"valarkovichMadHouseBot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const adminID int64 = 247753697

func main() {

	token := os.Getenv("tg_token")

	if token == "" {
		log.Fatal("tg_bot empty")
	}

	databaseURL := os.Getenv("DB_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is empty")
	}

	telegramBot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("cannot open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("cannot ping database: %v", err)
	}

	store := storage.New(db)
	if err := store.Init(); err != nil {
		log.Fatalf("cannot init storage: %v", err)
	}

	handler := appbot.New(telegramBot, store, adminID)

	log.Printf("authorized as @%s", telegramBot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := telegramBot.GetUpdatesChan(updateConfig)

	for update := range updates {
		handler.HandleUpdate(update)
	}
}
