package main

import (
	"flag"
	"log"
	"tgBotRecommender/clients/tgClient"
	"tgBotRecommender/consumer/eventConsumer"
	"tgBotRecommender/events/telegram"
	"tgBotRecommender/storage/files"
)

const (
	tgBotHost   = "api.telegram.org"
	storagePath = "storage"
	batchSize   = 100
)

func main() {
	eventsProcessor := telegram.New(
		tgClient.New(tgBotHost, mustToken()),
		files.NewStorage(storagePath),
	)

	log.Print("Starting bot")

	consumer := eventConsumer.NewConsumer(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(); err != nil {
		log.Fatal(err)
	}

}

func mustToken() string {
	token := flag.String("tg-bot-token",
		"",
		"provides access to tgClient bot")

	flag.Parse()

	if *token == "" {
		log.Fatal("token is required")
	}

	return *token
}
