package main

import (
	"log"
	"tgBotRecommender/clients/tgClient"
	"tgBotRecommender/consumer/eventConsumer"
	"tgBotRecommender/events/telegram"
	"tgBotRecommender/storage/files"
)

const (
	tgBotHost   = "api.telegram.org"
	storagePath = "files_storage"
	batchSize   = 100
	tgBotToken  = "7047428650:AAGnJCnA_RUZJ0TFntTYKqVYApD0vuQKNls"
)

func main() {
	eventsProcessor := telegram.New(
		tgClient.New(tgBotHost, tgBotToken),
		files.NewStorage(storagePath),
	)

	log.Print("Starting bot")

	consumer := eventConsumer.NewConsumer(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(); err != nil {
		log.Fatal(err)
	}
}
