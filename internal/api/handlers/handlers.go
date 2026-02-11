package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/aplication/usecases"
	vkkeyboard "github.com/LainIwakuras-father/Valentine-VK-Bot/internal/infra/vk"
	"github.com/SevereCloud/vksdk/v3/api"
)

// ValentineHandler обработчик валентинок
type ValentineHandler struct {
	vk           *api.VK
	service      *usecases.ValentineUseCases
	stateManager *StateManager
}

// NewValentineHandler создает новый обработчик валентинок
func NewValentineHandler(vk *api.VK, service *usecases.ValentineUseCases, stateManager *StateManager) *ValentineHandler {
	return &ValentineHandler{
		vk:           vk,
		service:      service,
		stateManager: stateManager,
	}
}

// Handle обрабатывает сообщения, связанные с валентинками
func (h *ValentineHandler) Handle(ctx context.Context, userID int, text string) bool {
	// Проверяем, есть ли активное состояние
	step, data := h.stateManager.GetState(userID)

	// Обрабатываем состояния
	switch step {
	case "waiting_anonymous":
		return h.handleAnonymous(ctx, userID, text)
	case "waiting_recipient":
		return h.handleRecipient(ctx, userID, text, data)
	case "waiting_valentine_type":
		return h.handleValentineType(ctx, userID, text, data)
	case "waiting_premade":
		return h.handlePremade(ctx, userID, text, data)
	case "waiting_custom_text":
		return h.handleCustomText(ctx, userID, text, data)
	}

	// Если нет состояния, проверяем команды
	switch text {
	case "💌 Отправить валентинку":
		h.startValentineSending(userID)
		return true
	case "📤 Мои отправленные":
		h.handleViewSent(ctx, userID)
		return true
	case "📥 Мои полученные":
		h.handleViewReceived(ctx, userID)
		return true
	}

	return false
}

// startValentineSending начинает процесс отправки валентинки
func (h *ValentineHandler) startValentineSending(userID int) {
	h.stateManager.SetState(userID, "waiting_anonymous")
	vkkeyboard.SendKeyboard(h.vk, userID,
		"Анонимная валентинка?",
		vkkeyboard.NewAnonymityKeyboard())
}

// handleAnonymous обрабатывает выбор анонимности
func (h *ValentineHandler) handleAnonymous(ctx context.Context, userID int, text string) bool {
	switch strings.ToLower(text) {
	case "да":
		h.stateManager.SetData(userID, "is_anonymous", true)
		h.stateManager.SetState(userID, "waiting_valentine_type")
		vkkeyboard.SendKeyboard(h.vk, userID,
			"Выберите тип валентинки:",
			vkkeyboard.NewValentineTypeKeyboard())
		return true
	case "нет":
		h.stateManager.SetData(userID, "is_anonymous", false)
		h.stateManager.SetState(userID, "waiting_recipient")
		vkkeyboard.SendMessage(h.vk, userID,
			"Введите ID получателя или ссылку на его страницу ВКонтакте:\n"+
				"Примеры: id123456789, https://vk.com/id123456789, @id123456789")
		return true
	default:
		vkkeyboard.SendKeyboard(h.vk, userID,
			"Пожалуйста, выберите 'Да' или 'Нет':",
			vkkeyboard.NewAnonymityKeyboard())
		return true
	}
}

// handleRecipient обрабатывает ввод получателя
func (h *ValentineHandler) handleRecipient(ctx context.Context, userID int, text string, data map[string]interface{}) bool {
	// Сохраняем ссылку получателя
	h.stateManager.SetData(userID, "recipient_link", text)
	h.stateManager.SetState(userID, "waiting_valentine_type")
	vkkeyboard.SendKeyboard(h.vk, userID,
		"Выберите тип валентинки:",
		vkkeyboard.NewValentineTypeKeyboard())
	return true
}

// handleValentineType обрабатывает выбор типа валентинки
func (h *ValentineHandler) handleValentineType(ctx context.Context, userID int, text string, data map[string]interface{}) bool {
	switch text {
	case "Заготовленная":
		h.stateManager.SetState(userID, "waiting_premade")
		vkkeyboard.SendKeyboard(h.vk, userID,
			"Выберите готовую валентинку:",
			vkkeyboard.NewPremadeImagesKeyboard())
		return true
	case "Собственная":
		h.stateManager.SetState(userID, "waiting_custom_text")
		vkkeyboard.SendMessage(h.vk, userID,
			"Введите текст вашей валентинки (максимум 500 символов):")
		return true
	default:
		vkkeyboard.SendKeyboard(h.vk, userID,
			"Выберите тип валентинки:",
			vkkeyboard.NewValentineTypeKeyboard())
		return true
	}
}

// handlePremade обрабатывает выбор готовой валентинки
func (h *ValentineHandler) handlePremade(ctx context.Context, userID int, text string, data map[string]interface{}) bool {
	// Определяем текст в зависимости от выбора
	var message, imageType string

	switch text {
	case "💝 Валентинка 1":
		message = "С Днем Святого Валентина! Ты делаешь этот мир лучше! ❤️"
		imageType = "premade_1"
	case "💘 Валентинка 2":
		message = "Ты - самое прекрасное, что случалось со мной! 💘"
		imageType = "premade_2"
	case "💖 Валентинка 3":
		message = "Мое сердце бьется только для тебя! 💖"
		imageType = "premade_3"
	case "💗 Валентинка 4":
		message = "Твоя улыбка - мое счастье! 💗"
		imageType = "premade_4"
	default:
		vkkeyboard.SendKeyboard(h.vk, userID,
			"Выберите готовую валентинку:",
			vkkeyboard.NewPremadeImagesKeyboard())
		return true
	}

	// Завершаем отправку
	h.finishValentineSending(ctx, userID, data, message, imageType)
	return true
}

// handleCustomText обрабатывает ввод текста своей валентинки
func (h *ValentineHandler) handleCustomText(ctx context.Context, userID int, text string, data map[string]interface{}) bool {
	// Проверяем длину текста
	if len(text) > 500 {
		vkkeyboard.SendMessage(h.vk, userID,
			"Текст слишком длинный (максимум 500 символов). Пожалуйста, введите короче:")
		return true
	}

	if len(text) < 3 {
		vkkeyboard.SendMessage(h.vk, userID,
			"Текст слишком короткий. Пожалуйста, введите сообщение хотя бы из 3 символов:")
		return true
	}

	// Завершаем отправку
	h.finishValentineSending(ctx, userID, data, text, "custom")
	return true
}

// finishValentineSending завершает отправку валентинки
func (h *ValentineHandler) finishValentineSending(ctx context.Context, userID int, data map[string]interface{}, message, imageType string) {
	// Получаем данные
	recipientLink, _ := data["recipient_link"].(string)
	isAnonymous, _ := data["is_anonymous"].(bool)

	// Если валентинка анонимная и ссылка не указана, используем ID пользователя как пример
	if isAnonymous && recipientLink == "" {
		recipientLink = fmt.Sprintf("id%d", userID+1) // Пример, нужно получить от пользователя
	}

	// Отправляем валентинку через сервис
	valentine, err := h.service.SendValentine(ctx, userID, recipientLink, message, isAnonymous, imageType)
	if err != nil {
		log.Printf("Ошибка отправки валентинки: %v", err)
		vkkeyboard.SendKeyboard(h.vk, userID,
			"❌ Ошибка при отправке валентинки: "+err.Error(),
			vkkeyboard.NewStartKeyboard())
	} else {
		// Формируем сообщение об успехе
		successMsg := "✅ Валентинка успешно создана!\n\n"
		if isAnonymous {
			successMsg += "Отправлено анонимно\n"
		} else {
			successMsg += "От вашего имени\n"
		}
		successMsg += fmt.Sprintf("Получатель: %s\n", recipientLink)
		successMsg += "📅 Валентинка будет доставлена 14 февраля!\n\n"
		successMsg += "Вы можете посмотреть свои отправленные валентинки в любое время."

		vkkeyboard.SendKeyboard(h.vk, userID, successMsg, vkkeyboard.NewStartKeyboard())

		log.Printf("Валентинка создана: ID=%s, отправитель=%d, получатель=%s",
			valentine.ID, userID, recipientLink)
	}

	// Очищаем состояние
	h.stateManager.ClearState(userID)
}

// handleViewSent обрабатывает просмотр отправленных валентинок
func (h *ValentineHandler) handleViewSent(ctx context.Context, userID int) {
	valentines, err := h.service.GetSentValentines(ctx, userID)
	if err != nil {
		log.Printf("Ошибка получения отправленных валентинок: %v", err)
		vkkeyboard.SendKeyboard(h.vk, userID,
			"❌ Произошла ошибка при получении ваших отправленных валентинок.",
			vkkeyboard.NewStartKeyboard())
		return
	}

	if len(valentines) == 0 {
		vkkeyboard.SendKeyboard(h.vk, userID,
			"📭 Вы еще не отправляли валентинок.",
			vkkeyboard.NewStartKeyboard())
		return
	}

	// Формируем сообщение
	message := "📤 Ваши отправленные валентинки:\n\n"

	for i, v := range valentines {
		// Форматируем статус отправки
		status := "⏳ Ожидает отправки 14 февраля"
		if v.SentAt != nil {
			status = fmt.Sprintf("✅ Отправлена %s", v.SentAt.Format("02.01.2006"))
		}

		// Форматируем анонимность
		//	anonymity := "👤 От вашего имени"
		//	if v.IsAnonymous {
		//		anonymity = "🎭 Анонимно"
		//	}

		message += fmt.Sprintf("%d. Для ID%d\n", i+1, v.RecipientID)
		message += fmt.Sprintf("   Сообщение: %s\n", v.FormatMessage())
		message += fmt.Sprintf("   %s | %s\n\n", anonymity, status)
	}

	vkkeyboard.SendKeyboard(h.vk, userID, message, vkkeyboard.NewStartKeyboard())
}

// handleViewReceived обрабатывает просмотр полученных валентинок
func (h *ValentineHandler) handleViewReceived(ctx context.Context, userID int) {
	// Проверяем, можно ли просматривать сегодня
	if !h.service.CanViewReceived() {
		vkkeyboard.SendKeyboard(h.vk, userID,
			"📅 Полученные валентинки можно посмотреть только 14 февраля!\n"+
				"Ждите этого дня, чтобы узнать, кто отправил вам валентинки! 💝",
			vkkeyboard.NewStartKeyboard())
		return
	}

	valentines, err := h.service.GetReceivedValentines(ctx, userID)
	if err != nil {
		log.Printf("Ошибка получения полученных валентинок: %v", err)
		vkkeyboard.SendKeyboard(h.vk, userID,
			"❌ Произошла ошибка при получении ваших полученных валентинок.",
			vkkeyboard.NewStartKeyboard())
		return
	}

	if len(valentines) == 0 {
		vkkeyboard.SendKeyboard(h.vk, userID,
			"📭 Вы еще не получали валентинок в этом году.",
			vkkeyboard.NewStartKeyboard())
		return
	}

	// Формируем сообщение
	message := "📥 Ваши полученные валентинки:\n\n"

	for i, v := range valentines {
		message += fmt.Sprintf("%d. От %s\n", i+1, v.GetSenderDisplay())
		message += fmt.Sprintf("   Сообщение: %s\n\n", v.Message)
	}

	message += "💖 С Днем Святого Валентина!"

	vkkeyboard.SendKeyboard(h.vk, userID, message, vkkeyboard.NewStartKeyboard())
}
