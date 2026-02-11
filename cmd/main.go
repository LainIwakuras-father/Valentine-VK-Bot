package main

import (
	"context"
	// 	"go/token"
	"log"
	"os"

	"github.com/joho/godotenv"

	// Импорт с точкой (тогда все функции будут доступны напрямую)
	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/infra/storage"
	keyboard "github.com/LainIwakuras-father/Valentine-VK-Bot/internal/infra/vk"

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
	log.Printf("Иницилизация Базы Данных...")
	// Инициализируем базу данных
	db, err := storage.NewSqliteDB()
	if err != nil {
		log.Fatal("Ошибка инициализации базы данных:", err)
	}
	defer func() {
		if err := storage.CloseDB(db); err != nil {
			log.Printf("Ошибка закрытия БД: %v", err)
		}
	}()
	// иницилизация repo

	vk := api.NewVK(token)
	log.Printf("Иницилизируем бота...")
	lp, err := longpoll.NewLongPoll(vk, 235791902)
	if err != nil {
		log.Fatal("Ошибка инициализации LongPoll:", err)
		panic(err)
	}

	// простой обработчик
	lp.MessageNew(func(ctx context.Context, obj events.MessageNewObject) {
		userID := obj.Message.PeerID
		text := obj.Message.Text

		switch text {
		case "Начать", "Привет", "Меню":
			// Отправляем главное меню
			keyboard.SendKeyboard(vk, userID, "Добро пожаловать в бот валентинок! Выберите действие:", keyboard.NewStartKeyboard())
		case "💌 Отправить валентинку":
			keyboard.SendKeyboard(vk, userID, "Анонимная валентинка?", keyboard.NewAnonymityKeyboard())
		case "Да", "Нет":
			keyboard.SendKeyboard(vk, userID, "Выберите тип валентинки:", keyboard.NewValentineTypeKeyboard())
		default:
			// Используем функцию из пакета vk
			keyboard.SendKeyboard(vk, userID, "Используйте кнопки меню",
				keyboard.NewStartKeyboard())

		}
	})

	log.Printf("Запускаем бота...")
	// Запуск
	if err := lp.Run(); err != nil {
		log.Fatal("Бот не смог запустится", err)
	}

	// Безопасное завершение
	// Ждет пока соединение закроется и события обработаются
	lp.Shutdown()

	// Закрыть соединение
	// Требует lp.Client.Transport = &http.Transport{DisableKeepAlives: true}
	lp.Client.CloseIdleConnections()
}
