package main

import (
	"context"
	// 	"go/token"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/SevereCloud/vksdk/v3/api"

	"github.com/SevereCloud/vksdk/v3/events"
	longpoll "github.com/SevereCloud/vksdk/v3/longpoll-bot"
)

func main() {
	// НЕ ЗАБЫВАЙ ПРО ПЕРЕМЕННЫЕ ОКРУЖЕНИЯ

	if err := godotenv.Load(); err != nil {
		log.Printf("Ошибка: .env file not found: %v", err)
	}
	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("Переменная окружения TOKEN не установлена!")
	}
	vk := api.NewVK(token)

	lp, err := longpoll.NewLongPoll(vk, 235791902)
	if err != nil {
		log.Fatal("Ошибка инициализации LongPoll:", err)
		panic(err)
	}

	lp.MessageNew(func(ctx context.Context, obj events.MessageNewObject) {
		log.Print(obj.Message.Text)
	})

	lp.Run()
}
