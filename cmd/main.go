package main

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"

	// Импорт с точкой (тогда все функции будут доступны напрямую)
	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/api/bot"
	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/aplication/usecases"
	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/infra/config"
	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/infra/storage"
	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/infra/storage/repositories"

	"github.com/SevereCloud/vksdk/v3/api"

	longpoll "github.com/SevereCloud/vksdk/v3/longpoll-bot"
)

func main() {
	log := config.SetupLogger()
	// НЕ ЗАБЫВАЙ ПРО ПЕРЕМЕННЫЕ ОКРУЖЕНИЯ

	if err := godotenv.Load(); err != nil {
		log.Info("Ошибка: .env file not found: %v", err)
	}
	token := os.Getenv("TOKEN")
	if token == "" {
		log.Error("Переменная окружения TOKEN не установлена!")
		os.Exit(1)
	}

	groupID := os.Getenv("GROUP_ID")
	if groupID == "" {
		log.Error("Переменная окружения GROUP_ID не установлена!")
		os.Exit(1)
	}
	// перевод в числовые значение строку с айдишником
	// 2. Преобразуем в int
	id, err := strconv.Atoi(groupID)
	if err != nil {
		log.Error("некорректный формат GROUP_ID: %w", err)
		os.Exit(1)
	}
	testMode := os.Getenv("TEST_MODE") == "true"
	log.Info("Иницилизация Базы Данных...")
	// Инициализируем базу данных
	db, err := storage.NewSqliteDB()
	if err != nil {
		log.Error("Ошибка инициализации базы данных:", err)
	}
	defer func() {
		if err := storage.CloseDB(db); err != nil {
			log.Error("Ошибка закрытия БД: %v", err)
		}
	}()
	// иницилизация repo
	repo := repositories.NewGORMValentineRepo(db)
	vk := api.NewVK(token)
	log.Info("Иницилизируем бота...")
	lp, err := longpoll.NewLongPoll(vk, id) // номер тестового сообщества 235791902
	if err != nil {
		log.Error("Ошибка инициализации LongPoll:", err)
		panic(err)
	}

	// простой обработчик
	botVk := bot.NewApp(vk, lp, repo, log, testMode)
	log.Info("Запускаем бота...")

	// планировщик отправлений валентинок
	scheduler := usecases.NewScheduler(vk, botVk.ValentineService, log)
	scheduler.Start()
	defer scheduler.Stop()

	// Запуск
	if err := botVk.Run(); err != nil {
		log.Error("Бот не смог запустится", err)
	}

	// Безопасное завершение
	// Ждет пока соединение закроется и события обработаются
	lp.Shutdown()

	// Закрыть соединение
	// Требует lp.Client.Transport = &http.Transport{DisableKeepAlives: true}
	lp.Client.CloseIdleConnections()
	log.Info("Бот завершил работу")
}
