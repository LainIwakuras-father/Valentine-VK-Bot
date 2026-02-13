package handlers

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
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
	if text == "❌ Отмена" || text == "отмена" || text == "Отмена" {
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
	case "waiting_custom_text":
		return h.handleCustomText(ctx, userID, text, data)
	case "waiting_photo_after_text":
		return h.handlePhotoAfterText(ctx, userID, text, data)
	case "waiting_photo_url":
		return h.handlePhotoURL(ctx, userID, text, data)
	case "waiting_custom_text_and_photo":
		return h.handleCustomTextAndPhoto(ctx, userID, text, attachments, data)
	case "waiting_premade_choice":
		return h.handlePremadeChoice(ctx, userID, text, data)
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
			"Примеры: id123456789, https://vk.com/id123456789, id123456789 (уберите символ @ из никнейма)")
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
		h.stateManager.SetState(userID, "waiting_premade_choice")

		// Сохраняем список attachment'ов в состояние пользователя
		h.stateManager.SetData(userID, "template_attachments", vkkeyboard.TemplateAttachments)

		// 1. Отправляем сообщение со всеми 5 фото
		attachments := strings.Join(vkkeyboard.TemplateAttachments, ",")
		if err := vkkeyboard.SendPhotoMessage(h.vk, userID,
			"🖼️ Вот доступные дизайны валентинок.\nВыберите номер понравившейся:",
			attachments); err != nil {
			h.log.Error("Ошибка отправки фото", "error", err)
		}

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

// выбор номера картинки
func (h *ValentineHandler) handlePremadeChoice(ctx context.Context, userID int, text string, data map[string]interface{}) bool {
	// Получаем сохранённый список attachment'ов
	raw, ok := data["template_attachments"]
	if !ok {
		h.log.Error("Не найден список attachment'ов", "user_id", userID)
		vkkeyboard.SendKeyboard(h.vk, userID, "❌ Ошибка, начните заново.", vkkeyboard.NewStartKeyboard())
		h.stateManager.ClearState(userID)
		return true
	}
	attachments, ok := raw.([]string)
	if !ok {
		h.log.Error("Неверный формат списка attachment'ов", "user_id", userID)
		vkkeyboard.SendKeyboard(h.vk, userID, "❌ Ошибка, начните заново.", vkkeyboard.NewStartKeyboard())
		h.stateManager.ClearState(userID)
		return true
	}

	// Парсим цифру
	index, err := strconv.Atoi(text)
	if err != nil || index < 1 || index > len(attachments) {
		vkkeyboard.SendKeyboard(h.vk, userID,
			fmt.Sprintf("❌ Введите цифру от 1 до %d:", len(attachments)),
			vkkeyboard.NewTemplateKeyboard())
		return true
	}

	// Выбранный attachment
	selected := attachments[index-1]

	// Стандартное сообщение (можно сделать разные под каждую картинку, если нужно)
	message := "С Днём Святого Валентина! ❤️"

	// Сохраняем валентинку
	h.finishValentineSending(ctx, userID, data, message, "template", selected)
	return true
}

// 5. Ввод собственного текста
func (h *ValentineHandler) handleCustomText(ctx context.Context, userID int, text string, data map[string]interface{}) bool {
	if len(text) > 500 {
		vkkeyboard.SendMessage(h.vk, userID, "❌ Текст слишком длинный (макс. 500 символов). Введите короче:")
		return true
	}
	if len(text) < 1 {
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
	if len(text) < 1 || len(text) > 500 {
		vkkeyboard.SendMessage(h.vk, userID, "❌ Текст должен быть от 1 до 500 символов. Попробуйте снова:")
		return true
	}

	// 2. Ищем фото во вложениях
	var photoAttachment string
	for _, att := range attachments {
		if att.Type == "photo" {
			//	h.log.Info("Получено фото-вложение", "attachment", original)
			// 🚀 ПЕРЕЗАЛИВАЕМ ФОТО
			newAttachment, err := h.reuploadUserPhoto(ctx, &att.Photo) // большой файл сучка надо лучше передавать указатель
			if err != nil {
				h.log.Error("Ошибка перезаливки фото", "error", err)
				// Сообщаем пользователю, но сохраняем валентинку без фото
				vkkeyboard.SendMessage(h.vk, userID, "⚠️ Не удалось обработать фото. Валентинка сохранена без фото.")
				photoAttachment = ""
			} else {
				photoAttachment = newAttachment
				h.log.Info("Фото перезалито", "new", photoAttachment)
			}
			break
		}
	}

	h.log.Info("Сохранение валентинки",
		"has_photo", photoAttachment != "",
		"photo_attachment", photoAttachment)
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
		successMsg += "Посмотреть отправленные можно в любой момент."
		vkkeyboard.SendKeyboard(h.vk, userID, successMsg, vkkeyboard.NewStartKeyboard())
		h.log.Info("Валентинка создана", "id", valentine.ID)

		// Уведомляем Пользователя о пришедщей валентинке
		if err := h.NotifyMassege(valentine.RecipientID); err != nil {
			h.log.Error("Ошибка уведомления",
				"recipient_id", valentine.RecipientID,
				"error", err)
		} else {
			h.log.Info("Уведомление отправлено",
				"recipient_id", valentine.RecipientID)
		}

		h.log.Info("Отправил уведомление")

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

	message := "📤 Вот все ваши отправленные валентинки!\n\n"
	for i, v := range valentines {
		// Формируем текст сообщения
		msg := fmt.Sprintf("📤 Отправленная валентинка #%d\n", i+1)
		msg += fmt.Sprintf("👤 Кому: %s\n", v.GetRecipientDisplay())
		msg += fmt.Sprintf("💌 Сообщение: %s\n", v.Message)
		if v.IsAnonymous {
			msg += "🎭 Анонимно\n"
		} else {
			msg += "👤 От вашего имени\n"
		}

		// если с фото то оправить с фото если нет то обычное сообщение
		if v.PhotoURL != "" {
			vkkeyboard.SendPhotoMessage(h.vk, userID, msg, v.PhotoURL)
		} else if err = vkkeyboard.SendMessage(h.vk, userID, msg); err != nil {
			h.log.Error("Ошибка отправки сообщения с валентинкой",
				"valentine_id", v.ID, "error", err)
		}
		// Небольшая задержка, чтобы не флудить
		time.Sleep(300 * time.Millisecond)

	}

	vkkeyboard.SendKeyboard(h.vk, userID, message, vkkeyboard.NewStartKeyboard())
}

// ------------------- ПРОСМОТР ПОЛУЧЕННЫХ -------------------

func (h *ValentineHandler) handleViewReceived(ctx context.Context, userID int) {
	if !h.service.CanViewReceived() {
		msg := fmt.Sprintf("📅 Полученные валентинки можно посмотреть только с 14 февраля!")
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

	msg := "📥 вот все ваши полученные валентинки!\n\n"
	for i, v := range valentines {
		msg := fmt.Sprintf("📥 Полученная валентинка #%d\n", i+1)
		msg += fmt.Sprintf("🎁 От: %s\n", v.GetSenderDisplay())
		msg += fmt.Sprintf("💌 %s\n", v.Message)
		h.log.Info("Вот такой урл фото", "URL", v.PhotoURL)
		// если с фото то оправить с фото если нет то обычное сообщение
		if v.PhotoURL != "" {
			vkkeyboard.SendPhotoMessage(h.vk, userID, msg, v.PhotoURL)
		} else if err = vkkeyboard.SendMessage(h.vk, userID, msg); err != nil {
			h.log.Error("Ошибка отправки сообщения с валентинкой",
				"valentine_id", v.ID, "error", err)
		}
		// Небольшая задержка, чтобы не флудить
		time.Sleep(300 * time.Millisecond)

	}
	msg += "💖 С Днём Святого Валентина!"
	vkkeyboard.SendKeyboard(h.vk, userID, msg, vkkeyboard.NewStartKeyboard())
}

// ------------------- СТАРТ ОТПРАВКИ -------------------

func (h *ValentineHandler) startValentineSending(userID int) {
	h.stateManager.SetState(userID, "waiting_anonymous")
	vkkeyboard.SendKeyboard(h.vk, userID,
		"Анонимная валентинка?",
		vkkeyboard.NewAnonymityKeyboard())
}

// --------УВЕДОМЛЕНИЕ ПОЛУЧАТЕЛЯ
func (h *ValentineHandler) NotifyMassege(recipientID int) error { // возможно добавить слова про то какая валентинка.. хз
	notify := "💝 Вам отправлена валентинка💝 \n\n"
	notify += "Посмотреть можно нажавь кнопку '📥 Мои полученные'!\n\n"

	err := vkkeyboard.SendKeyboard(h.vk, recipientID, notify, vkkeyboard.NewStartKeyboard())
	if err != nil {
		h.log.Error("ошибка отправки уведомления", "error", err)
		return fmt.Errorf("ошибка отправки уведомления пользователю %d: %w", recipientID, err)
	}
	return nil
}

// --------------ЗАГРУЗКА ФОТО ------
// reuploadUserPhoto скачивает фото по attachment, загружает на сервер сообщений и возвращает новый attachment
func (h *ValentineHandler) reuploadUserPhoto(ctx context.Context, photo *object.PhotosPhoto) (string, error) {
	h.log.Info("Начинаем перезаливку фото")

	// 1. Получаем URL оригинального фото
	// 1. Получаем URL самого большого размера
	if len(photo.Sizes) == 0 {
		return "", fmt.Errorf("нет доступных размеров фото")
	}
	largest := photo.Sizes[len(photo.Sizes)-1]
	photoURL := largest.URL
	h.log.Info("Скачиваем фото", "url", photoURL)

	// 2. Скачиваем фото
	resp, err := http.Get(photoURL)
	if err != nil {
		return "", fmt.Errorf("ошибка скачивания: %w", err)
	}
	defer resp.Body.Close()
	photoBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения: %w", err)
	}

	// 3. Получаем сервер для загрузки в сообщения
	uploadServer, err := h.vk.PhotosGetMessagesUploadServer(api.Params{})
	if err != nil {
		return "", fmt.Errorf("ошибка получения upload server: %w", err)
	}

	// 4. Загружаем фото на сервер
	uploadResp, err := vkkeyboard.UploadPhotoToServer(uploadServer.UploadURL, photoBytes)
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки на сервер: %w", err)
	}

	// 5. Сохраняем фото в сообществе
	savedPhotos, err := h.vk.PhotosSaveMessagesPhoto(api.Params{
		"photo":  uploadResp.Photo,
		"server": uploadResp.Server,
		"hash":   uploadResp.Hash,
	})
	if err != nil {
		return "", fmt.Errorf("ошибка сохранения фото: %w", err)
	}
	if len(savedPhotos) == 0 {
		return "", fmt.Errorf("фото не сохранилось")
	}

	// 6. Формируем новый attachment
	newAttachment := fmt.Sprintf("photo%d_%d", savedPhotos[0].OwnerID, savedPhotos[0].ID)
	h.log.Info("Фото успешно перезалито",
		"old", fmt.Sprintf("photo%d_%d", photo.OwnerID, photo.ID),
		"new", newAttachment,
		"owner_id", savedPhotos[0].OwnerID)
	return newAttachment, nil
}

// getPhotoURLByAttachment — получает прямую ссылку на фото (самый большой размер)
func (h *ValentineHandler) getPhotoURLByAttachment(attachment string) (string, error) {
	trimmed := strings.TrimPrefix(attachment, "photo")
	parts := strings.Split(trimmed, "_")
	if len(parts) != 2 {
		return "", fmt.Errorf("неверный формат attachment")
	}
	ownerID := parts[0]
	photoID := parts[1]

	photos, err := h.vk.PhotosGetByID(api.Params{
		"photos": fmt.Sprintf("%s_%s", ownerID, photoID),
	})
	if err != nil {
		return "", err
	}
	if len(photos) == 0 {
		return "", fmt.Errorf("фото не найдено")
	}
	photo := photos[0]
	if len(photo.Sizes) == 0 {
		return "", fmt.Errorf("нет размеров фото")
	}
	// Берем последний (обычно самый большой) размер
	largest := photo.Sizes[len(photo.Sizes)-1]
	return largest.URL, nil
}
