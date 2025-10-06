package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"tgBotRecommender/clients/tgClient"
	"tgBotRecommender/consumer/eventConsumer"
	"tgBotRecommender/events/telegram"
	"tgBotRecommender/storage/database"
)

const (
	tgBotHost = "api.telegram.org"
	batchSize = 100
)

func main() {
	// Инициализируем базу данных
	db, err := database.HandleConn()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	eventsProcessor := telegram.New(
		tgClient.New(tgBotHost, mustToken()),
		database.Storage{},
	)

	log.Print("Starting bot")

	// Запускаем HTTP сервер для Render health checks
	go startHTTPServer()

	consumer := eventConsumer.NewConsumer(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(); err != nil {
		log.Fatal(err)
	}
}

func startHTTPServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Telegram Bot is running"))
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Render автоматически устанавливает PORT переменную
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting HTTP server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Printf("HTTP server error: %v", err)
	}
}

func mustToken() string {
	if token := os.Getenv("TG_BOT_TOKEN"); token != "" {
		return token
	}

	token := flag.String("tg-bot-token", "", "provides access to tgClient bot")
	flag.Parse()

	if *token == "" {
		log.Fatal("token is required (set TG_BOT_TOKEN env var or use --tg-bot-token flag)")
	}

	return *token
}
