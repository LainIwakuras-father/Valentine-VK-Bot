package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	// Импорт с точкой (тогда все функции будут доступны напрямую)
	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/api/bot"
	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/infra/storage"
	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/infra/storage/repositories"

	"github.com/SevereCloud/vksdk/v3/api"

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
	repo := repositories.NewGORMValentineRepo(db)
	vk := api.NewVK(token)
	log.Printf("Иницилизируем бота...")
	lp, err := longpoll.NewLongPoll(vk, 235791902)
	if err != nil {
		log.Fatal("Ошибка инициализации LongPoll:", err)
		panic(err)
	}

	// простой обработчик
	botVk := bot.NewApp(vk, lp, repo)
	log.Printf("Запускаем бота...")
	// Запуск
	if err := botVk.Run(); err != nil {
		log.Fatal("Бот не смог запустится", err)
	}

	// Безопасное завершение
	// Ждет пока соединение закроется и события обработаются
	lp.Shutdown()

	// Закрыть соединение
	// Требует lp.Client.Transport = &http.Transport{DisableKeepAlives: true}
	lp.Client.CloseIdleConnections()
	log.Println("Бот завершил работу")
}
