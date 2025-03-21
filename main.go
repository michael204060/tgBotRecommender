package main

import (
	"flag"
	"log"
	"tgBotRecommender/clients/tgClient"
	"tgBotRecommender/consumer/eventConsumer"
	"tgBotRecommender/events/telegram"
	"tgBotRecommender/storage/database"
)

const (
	tgBotHost = "api.telegram.org"
	batchSize = 10000
)

func main() {
	eventsProcessor := telegram.New(
		tgClient.New(tgBotHost, mustToken()),
		//database.NewStorage(connStr), //>???
		database.Storage{},
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
