package eventConsumer

import (
	"log"
	"tgBotRecommender/events"
	"time"
)

type Consumer struct {
	fetcher   events.Fetcher
	processor events.Processor
	batchSize int
}

func NewConsumer(fetcher events.Fetcher, processor events.Processor, batchSize int) Consumer {
	return Consumer{
		fetcher:   fetcher,
		processor: processor,
		batchSize: batchSize,
	}
}

func (cons Consumer) Start() error {
	for {
		gotEvents, err := cons.fetcher.Fetch(cons.batchSize)
		if err != nil {
			log.Printf("[ERROR] Failed to fetch events: %s", err.Error())
			continue
		}

		if len(gotEvents) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		if err := cons.handleEvents(gotEvents); err != nil {
			log.Print(err)
			continue
		}
	}
}

func (cons *Consumer) handleEvents(events []events.Event) error {
	for _, event := range events {
		log.Printf("received new event: %s", event.Text)

		if err := cons.processor.Process(event); err != nil {
			log.Printf("failed to handle event: %s", err.Error())
			continue
		}
	}
	return nil
}
