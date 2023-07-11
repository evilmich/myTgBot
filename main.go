package main

import (
	"log"
	tgClient "my_tg_bot/clients/telegram"
	"my_tg_bot/config"
	"my_tg_bot/consumer/event_consumer"
	tgProcess "my_tg_bot/events/telegram"
	"my_tg_bot/storage/mongodb"
	"time"
)

const (
	tgBotHost = "api.telegram.org"
	//storagePath = "files_storage"
	batchSize = 100
)

func main() {
	cfg := config.MustLoad()
	//storage := files.New(storagePath)

	storage := mongodb.New(cfg.MongoConnectionString, 10*time.Second)

	eventsProcessor := tgProcess.New(
		tgClient.New(tgBotHost, cfg.TgBotToken),
		storage,
	)

	log.Print("service started")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(); err != nil {
		log.Fatal("service is stopped", err)
	}
}
