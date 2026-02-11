package bot

import (
	"context"
	"log"

	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/aplication/usecases"
	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/infra/storage/repositories"
	vkkeyboard "github.com/LainIwakuras-father/Valentine-VK-Bot/internal/infra/vk"
	"github.com/SevereCloud/vksdk/v3/api"
	"github.com/SevereCloud/vksdk/v3/events"
	"github.com/SevereCloud/vksdk/v3/longpoll-bot"
)

// App представляет основное приложение бота
type App struct {
	vk               *api.VK
	lp               *longpoll.LongPoll
	valentineService *usecases.ValentineUseCases
	stateManager     *state.StateManager
	valentineHandler *handlers.ValentineHandler
}

// NewApp создает новый экземпляр приложения
func NewApp(vk *api.VK, lp *longpoll.LongPoll, repo repositories.GORMValentineRepository) *App {
	// Создаем сервисы
	valentineService := usecases.NewValentineUseCases(repo)
	stateManager := usecases.NewStateManager()
	valentineHandler := handlers.NewValentineHandler(vk, valentineService, stateManager)

	return &App{
		vk:               vk,
		lp:               lp,
		valentineService: valentineService,
		stateManager:     stateManager,
		valentineHandler: valentineHandler,
	}
}

// Run запускает бота
func (app *App) Run() error {
	// Регистрируем обработчики
	app.registerHandlers()

	log.Printf("Запускаем бота...")
	return app.lp.Run()
}

// registerHandlers регистрирует обработчики событий
func (app *App) registerHandlers() {
	app.lp.MessageNew(func(ctx context.Context, obj events.MessageNewObject) {
		app.handleMessage(ctx, obj)
	})
}

// handleMessage обрабатывает входящие сообщения
func (app *App) handleMessage(ctx context.Context, obj events.MessageNewObject) {
	userID := obj.Message.PeerID
	text := obj.Message.Text

	// Обработка приветственных команд
	if text == "Начать" || text == "Привет" || text == "Меню" {
		vkkeyboard.SendKeyboard(app.vk, userID,
			"Добро пожаловать в бот валентинок! 💝\n"+
				"Здесь вы можете отправлять и получать валентинки.\n\n"+
				"Как это работает:\n"+
				"1. Отправьте валентинку - она сохранится и будет доставлена 14 февраля\n"+
				"2. Посмотрите свои отправленные валентинки в любое время\n"+
				"3. Посмотрите полученные валентинки 14 февраля\n\n"+
				"Выберите действие:",
			vkkeyboard.NewStartKeyboard())
		return
	}

	// Пробуем обработать через ValentineHandler
	if app.valentineHandler.Handle(ctx, userID, text) {
		return
	}

	// Если не обработано, показываем главное меню
	vkkeyboard.SendKeyboard(app.vk, userID,
		"Используйте кнопки меню для навигации",
		vkkeyboard.NewStartKeyboard())
}
