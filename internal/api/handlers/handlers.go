package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	//	"strconv"

	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/aplication/usecases"
	vkkeyboard "github.com/LainIwakuras-father/Valentine-VK-Bot/internal/infra/vk"
	"github.com/SevereCloud/vksdk/v3/api"
	"github.com/SevereCloud/vksdk/v3/events"
	"github.com/SevereCloud/vksdk/v3/object"
)

// ValentineHandler обработчик валентинок
type ValentineHandler struct {
	vk           *api.VK
	service      *usecases.ValentineUseCases
	stateManager *StateManager
	log          *slog.Logger
}

// NewValentineHandler создаёт новый обработчик
func NewValentineHandler(vk *api.VK, service *usecases.ValentineUseCases, stateManager *StateManager, log *slog.Logger) *ValentineHandler {
	return &ValentineHandler{
		vk:           vk,
		service:      service,
		stateManager: stateManager,
		log:          log.With("component", "valentine_handler"),
	}
}

// ------------------- СОСТОЯНИЯ -------------------

// Handle обрабатывает сообщения
func (h *ValentineHandler) Handle(ctx context.Context, obj events.MessageNewObject) bool {
	userID := obj.Message.PeerID
	text := obj.Message.Text
	attachments := obj.Message.Attachments

	// Глобальная отмена
	if text == "❌ Отмена" {
		h.stateManager.ClearState(userID)
		vkkeyboard.SendKeyboard(h.vk, userID, "❌ Отправка отменена.", vkkeyboard.NewStartKeyboard())
		return true
	}

	step, data := h.stateManager.GetState(userID)
	h.log.Debug("Обработка", "user_id", userID, "text", text, "step", step)

	switch step {
	case "waiting_anonymous":
		return h.handleAnonymous(ctx, userID, text, data)
	case "waiting_recipient":
		return h.handleRecipient(ctx, userID, text, data)
	case "waiting_valentine_type":
		return h.handleValentineType(ctx, userID, text, data)
	case "waiting_premade":
		return h.handlePremade(ctx, userID, text, data)
	case "waiting_custom_text":
		return h.handleCustomText(ctx, userID, text, data)
	case "waiting_photo_after_text":
		return h.handlePhotoAfterText(ctx, userID, text, data)
	case "waiting_photo_url":
		return h.handlePhotoURL(ctx, userID, text, data)
	case "waiting_custom_text_and_photo":
		return h.handleCustomTextAndPhoto(ctx, userID, text, attachments, data)
	}

	// Команды без состояния
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
	case "test_send_all":
		h.handleTestSendAll(ctx, userID)
		return true
	}
	return false
}

// ------------------ СОСТОЯНИЯ ------------------

// 1. Анонимность
func (h *ValentineHandler) handleAnonymous(ctx context.Context, userID int, text string, data map[string]interface{}) bool {
	switch strings.ToLower(text) {
	case "да":
		h.stateManager.SetData(userID, "is_anonymous", true)
	case "нет":
		h.stateManager.SetData(userID, "is_anonymous", false)
	default:
		vkkeyboard.SendKeyboard(h.vk, userID, "Выберите 'Да' или 'Нет':", vkkeyboard.NewAnonymityKeyboard())
		return true
	}
	// Переходим к вводу получателя
	h.stateManager.SetState(userID, "waiting_recipient")
	vkkeyboard.SendMessage(h.vk, userID,
		"Введите ID или ссылку на профиль ВКонтакте получателя:\n"+
			"Примеры: id123456789, https://vk.com/id123456789, @id123456789")
	return true
}

// 2. Получатель
func (h *ValentineHandler) handleRecipient(ctx context.Context, userID int, text string, data map[string]interface{}) bool {
	// Сохраняем ссылку как есть, парсинг будет в usecase
	h.stateManager.SetData(userID, "recipient_link", text)
	h.stateManager.SetState(userID, "waiting_valentine_type")
	vkkeyboard.SendKeyboard(h.vk, userID, "Выберите тип валентинки:", vkkeyboard.NewValentineTypeKeyboard())
	return true
}

// 3. Тип валентинки
func (h *ValentineHandler) handleValentineType(ctx context.Context, userID int, text string, data map[string]interface{}) bool {
	switch text {
	case "Заготовленная":
		h.stateManager.SetState(userID, "waiting_premade")
		vkkeyboard.SendKeyboard(h.vk, userID, "Выберите готовую валентинку:", vkkeyboard.NewTemplateKeyboard())
		return true
	case "Собственная":
		h.stateManager.SetState(userID, "waiting_custom_text_and_photo")
		vkkeyboard.SendMessage(h.vk, userID,
			"✍️ Напишите текст валентинки и **прикрепите фото** (необязательно).\n"+
				"Отправьте одним сообщением: текст + вложение.")
		return true
	default:
		vkkeyboard.SendKeyboard(h.vk, userID, "Выберите тип:", vkkeyboard.NewValentineTypeKeyboard())
		return true
	}
}

// Предопределённые attachment'ы готовых валентинок
var templateAttachments = map[string]string{
	"💝 1": "photo-123456_789012", // замените на реальные ID фото из вашего сообщества
	"💘 2": "photo-123456_789013",
	"💖 3": "photo-123456_789014",
	"💗 4": "photo-123456_789015",
}

func (h *ValentineHandler) handlePremade(ctx context.Context, userID int, text string, data map[string]interface{}) bool {
	// Если текст — одна из кнопок шаблона
	if attachment, ok := templateAttachments[text]; ok {
		// Берём стандартное сообщение для этого шаблона
		message := "С Днём Святого Валентина! ❤️"
		h.finishValentineSending(ctx, userID, data, message, "template", attachment)
		return true
	}

	// Иначе показываем клавиатуру шаблонов
	h.stateManager.SetState(userID, "waiting_premade")
	vkkeyboard.SendKeyboard(h.vk, userID,
		"Выберите дизайн валентинки:",
		vkkeyboard.NewTemplateKeyboard())
	return true
}

// 5. Ввод собственного текста
func (h *ValentineHandler) handleCustomText(ctx context.Context, userID int, text string, data map[string]interface{}) bool {
	if len(text) > 500 {
		vkkeyboard.SendMessage(h.vk, userID, "❌ Текст слишком длинный (макс. 500 символов). Введите короче:")
		return true
	}
	if len(text) < 3 {
		vkkeyboard.SendMessage(h.vk, userID, "❌ Текст слишком короткий. Введите хотя бы 3 символа:")
		return true
	}
	// Сохраняем текст
	h.stateManager.SetData(userID, "custom_text", text)
	// Предлагаем добавить фото
	h.stateManager.SetState(userID, "waiting_photo_after_text")
	vkkeyboard.SendKeyboard(h.vk, userID, "Хотите добавить фото к валентинке?", vkkeyboard.NewPhotoAfterTextKeyboard())
	return true
}

// 6. Добавление фото после текста
func (h *ValentineHandler) handlePhotoAfterText(ctx context.Context, userID int, text string, data map[string]interface{}) bool {
	switch text {
	case "📷 Добавить фото":
		h.stateManager.SetState(userID, "waiting_photo_url")
		vkkeyboard.SendMessage(h.vk, userID,
			"Отправьте прямую ссылку на фото (JPG, PNG, GIF).\n"+
				"Или напишите 'пропустить'.")
		return true
	case "⏭ Отправить без фото":
		// Завершаем без фото
		customText, _ := data["custom_text"].(string)
		h.finishValentineSending(ctx, userID, data, customText, "custom", "")
		return true
	default:
		vkkeyboard.SendKeyboard(h.vk, userID, "Выберите действие:", vkkeyboard.NewPhotoAfterTextKeyboard())
		return true
	}
}

// 7. Ввод URL фото
func (h *ValentineHandler) handlePhotoURL(ctx context.Context, userID int, text string, data map[string]interface{}) bool {
	if strings.ToLower(text) == "пропустить" {
		customText, _ := data["custom_text"].(string)
		h.finishValentineSending(ctx, userID, data, customText, "custom", "")
		return true
	}

	if !vkkeyboard.IsValidPhotoURL(text) {
		vkkeyboard.SendMessage(h.vk, userID,
			"❌ Некорректная ссылка. Отправьте прямую ссылку на фото или 'пропустить'.")
		return true
	}

	customText, _ := data["custom_text"].(string)
	h.finishValentineSending(ctx, userID, data, customText, "custom", text)
	return true
}

// Новый обработчик
func (h *ValentineHandler) handleCustomTextAndPhoto(ctx context.Context, userID int, text string, attachments []object.MessagesMessageAttachment, data map[string]interface{}) bool {
	// 1. Проверяем текст
	if len(text) < 3 || len(text) > 500 {
		vkkeyboard.SendMessage(h.vk, userID, "❌ Текст должен быть от 3 до 500 символов. Попробуйте снова:")
		return true
	}

	// 2. Ищем фото во вложениях
	var photoAttachment string
	for _, att := range attachments {
		if att.Type == "photo" {
			photoAttachment = fmt.Sprintf("photo%d_%d", att.Photo.OwnerID, att.Photo.ID)
			h.log.Info("Получено фото-вложение", "attachment", photoAttachment)
			break
		}
	}

	// 3. Если фото нет, просто продолжаем без него
	h.finishValentineSending(ctx, userID, data, text, "custom", photoAttachment)
	return true
}

// ------------------- ЗАВЕРШЕНИЕ ОТПРАВКИ -------------------

// finishValentineSending сохраняет валентинку в БД и завершает процесс

func (h *ValentineHandler) finishValentineSending(ctx context.Context, userID int, data map[string]interface{}, message, imageType, photoURL string) {
	recipientLink, _ := data["recipient_link"].(string)
	isAnonymous, _ := data["is_anonymous"].(bool)

	if recipientLink == "" {
		h.log.Error("Нет получателя", "user_id", userID)
		vkkeyboard.SendKeyboard(h.vk, userID, "❌ Ошибка: получатель не указан. Начните заново.", vkkeyboard.NewStartKeyboard())
		h.stateManager.ClearState(userID)
		return
	}

	// Определяем финальный тип: если передана photoURL — значит будет фото
	finalImageType := imageType
	if photoURL != "" {
		finalImageType = "photo"
	}

	valentine, err := h.service.SendValentine(ctx, userID, recipientLink, message, isAnonymous, finalImageType, photoURL)
	if err != nil {
		h.log.Error("Ошибка создания валентинки", "error", err)
		vkkeyboard.SendKeyboard(h.vk, userID, "❌ Ошибка: "+err.Error(), vkkeyboard.NewStartKeyboard())
	} else {
		now := time.Now()
		successMsg := "✅ Валентинка успешно создана!\n\n"
		if isAnonymous {
			successMsg += "🎭 Анонимная\n"
		} else {
			successMsg += "👤 От вашего имени\n"
		}
		successMsg += fmt.Sprintf("📨 Получатель: %s\n", recipientLink)
		if photoURL != "" {
			successMsg += "📷 С фото\n"
		}
		if now.Month() == time.February && now.Day() == 14 {
			successMsg += "🎉 Отправлена немедленно (сегодня 14 февраля)!\n\n"
		} else {
			successMsg += "📅 Будет доставлена 14 февраля!\n\n"
		}
		successMsg += "Посмотреть отправленные можно в любой момент."
		vkkeyboard.SendKeyboard(h.vk, userID, successMsg, vkkeyboard.NewStartKeyboard())
		h.log.Info("Валентинка создана", "id", valentine.ID)
	}
	h.stateManager.ClearState(userID)
}

// ------------------- ПРОСМОТР ОТПРАВЛЕННЫХ -------------------

func (h *ValentineHandler) handleViewSent(ctx context.Context, userID int) {
	valentines, err := h.service.GetSentValentines(ctx, userID)
	if err != nil {
		h.log.Error("Ошибка получения отправленных", "user_id", userID, "error", err)
		vkkeyboard.SendKeyboard(h.vk, userID,
			"❌ Не удалось получить список отправленных валентинок.",
			vkkeyboard.NewStartKeyboard())
		return
	}

	if len(valentines) == 0 {
		vkkeyboard.SendKeyboard(h.vk, userID,
			"📭 Вы ещё не отправляли валентинок.",
			vkkeyboard.NewStartKeyboard())
		return
	}

	message := "📤 Ваши отправленные валентинки:\n\n"
	for i, v := range valentines {
		status := "⏳ Ожидает 14 февраля"
		if v.IsSent() {
			status = fmt.Sprintf("✅ Отправлена %s", v.SentAt.Format("02.01.2006"))
		}
		anon := "👤 Открыто"
		if v.IsAnonymous {
			anon = "🎭 Анонимно"
		}
		message += fmt.Sprintf("%d. Для ID%d\n", i+1, v.RecipientID)
		message += fmt.Sprintf("   💌 %s\n", v.FormatMessage())
		message += fmt.Sprintf("   %s | %s\n\n", anon, status)
	}

	sent, received, _ := h.service.GetStats(ctx, userID)
	message += fmt.Sprintf("📊 Статистика: отправлено %d, получено %d", sent, received)

	vkkeyboard.SendKeyboard(h.vk, userID, message, vkkeyboard.NewStartKeyboard())
}

// ------------------- ПРОСМОТР ПОЛУЧЕННЫХ -------------------

func (h *ValentineHandler) handleViewReceived(ctx context.Context, userID int) {
	if !h.service.CanViewReceived() {
		//	now := time.Now()
		//	next := time.Date(now.Year()+1, time.February, 14, 0, 0, 0, 0, now.Location())
		//	days := int(next.Sub(now).Hours() / 24)
		msg := fmt.Sprintf("📅 Полученные валентинки можно посмотреть только с 14 февраля!")
		// ⏳ Осталось %d дней."), days)
		vkkeyboard.SendKeyboard(h.vk, userID, msg, vkkeyboard.NewStartKeyboard())
		return
	}

	valentines, err := h.service.GetReceivedValentines(ctx, userID)
	h.log.Info("пользователь", userID)
	h.log.Info("Валентинок найдено для пользователя:", valentines)
	if err != nil {
		h.log.Error("Ошибка получения полученных", "user_id", userID, "error", err)
		vkkeyboard.SendKeyboard(h.vk, userID,
			"❌ Не удалось получить полученные валентинки.",
			vkkeyboard.NewStartKeyboard())
		return
	}

	if len(valentines) == 0 {
		vkkeyboard.SendKeyboard(h.vk, userID,
			"📭 В этом году вы ещё не получали валентинок.\nНо они ещё могут прийти! 💘",
			vkkeyboard.NewStartKeyboard())
		return
	}

	msg := "📥 Ваши полученные валентинки:\n\n"
	for i, v := range valentines {
		msg += fmt.Sprintf("%d. От %s\n", i+1, v.GetSenderDisplay())
		msg += fmt.Sprintf("   💌 %s\n", v.Message)
		if v.PhotoURL != "" {
			msg += "   📷 С фото\n"
		}
		if v.SentAt != nil {
			msg += fmt.Sprintf("   🕐 %s\n\n", v.SentAt.Format("02.01.2006"))
		}
	}
	msg += "💖 С Днём Святого Валентина!"
	vkkeyboard.SendKeyboard(h.vk, userID, msg, vkkeyboard.NewStartKeyboard())
}

// ------------------- ТЕСТОВАЯ КОМАНДА (для админов) -------------------

func (h *ValentineHandler) handleTestSendAll(ctx context.Context, userID int) {
	h.log.Info("Ручная отправка всех валентинок", "initiated_by", userID)

	valentines, err := h.service.GetUnsentValentines(ctx)
	if err != nil {
		h.log.Error("Ошибка получения неотправленных", "error", err)
		vkkeyboard.SendMessage(h.vk, userID, "❌ Ошибка получения валентинок")
		return
	}

	if len(valentines) == 0 {
		vkkeyboard.SendMessage(h.vk, userID, "✅ Нет неотправленных валентинок")
		return
	}

	// Здесь можно добавить реальную отправку, но пока просто помечаем как отправленные
	for _, v := range valentines {
		_ = h.service.MarkValentineAsSent(ctx, v.ID) // игнорируем ошибку для демо
	}

	vkkeyboard.SendMessage(h.vk, userID,
		fmt.Sprintf("✅ Помечено как отправлено: %d валентинок", len(valentines)))
}

// ------------------- СТАРТ ОТПРАВКИ -------------------

func (h *ValentineHandler) startValentineSending(userID int) {
	h.stateManager.SetState(userID, "waiting_anonymous")
	vkkeyboard.SendKeyboard(h.vk, userID,
		"Анонимная валентинка?",
		vkkeyboard.NewAnonymityKeyboard())
}
