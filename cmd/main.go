package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"time"
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

	databaseURL := os.Getenv("DB_URLexport")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is empty")
	}

	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, "tcp4", addr)
		},
		ForceAttemptHTTP2:     false,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 0,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Timeout:   70 * time.Second,
		Transport: transport,
	}

	telegramBot, err := tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, client)
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
