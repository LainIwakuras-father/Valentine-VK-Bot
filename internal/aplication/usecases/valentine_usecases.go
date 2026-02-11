package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/domain"
	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/infra/storage/repositories"
	"github.com/SevereCloud/vksdk/v3/api"
	"github.com/google/uuid"
)

type ValentineUseCases struct {
	repo     repositories.ValentineRepository
	vk       *api.VK // добавляем
	log      *slog.Logger
	testMode bool
}

// Конструктор теперь принимает *api.VK
func NewValentineUseCases(repo repositories.ValentineRepository, vk *api.VK, log *slog.Logger, testMode bool) *ValentineUseCases {
	return &ValentineUseCases{
		repo:     repo,
		vk:       vk,
		log:      log.With("component", "valentine_usecases"),
		testMode: testMode,
	}
}

// SendValentine отправляет валентинку
func (uc *ValentineUseCases) SendValentine(
	ctx context.Context,
	senderID int,
	recipientLink string,
	message string,
	isAnonymous bool,
	imageType string,
	photoURL string,
) (*domain.Valentine, error) {
	uc.log.Debug("Отправка валентинки",
		"sender_id", senderID,
		"recipient_link", recipientLink,
		"is_anonymous", isAnonymous,
		"image_type", imageType)

	// Парсим ID получателя из ссылки
	recipientID, recipientOriginal, err := uc.parseRecipient(recipientLink)
	if err != nil {
		uc.log.Error("Ошибка парсинга получателя", "error", err)
		return nil, fmt.Errorf("ошибка парсинга получателя: %w", err)
	}
	// Проверка, что ID не нулевой (на случай, если parseRecipient вернул 0 без ошибки)
	if recipientID == 0 {
		return nil, fmt.Errorf("не удалось определить ID получателя")
	}

	// Проверяем, что отправитель не отправляет себе
	if senderID == recipientID {
		err := fmt.Errorf("нельзя отправить валентинку самому себе")
		uc.log.Error("Попытка отправить валентинку себе", "error", err)
		return nil, err
	}

	// Создаем валентинку
	valentine := &domain.Valentine{
		ID:                uuid.New().String(),
		SenderID:          senderID,
		RecipientID:       recipientID,
		RecipientOriginal: recipientOriginal,
		Message:           message,
		ImageType:         imageType,
		PhotoURL:          photoURL,
		IsAnonymous:       isAnonymous,
		Opened:            false,
	}

	// Сохраняем в БД
	if err := uc.repo.Create(ctx, valentine); err != nil {
		uc.log.Error("Ошибка сохранения валентинки", "error", err)
		return nil, fmt.Errorf("ошибка сохранения валентинки: %w", err)
	}

	uc.log.Info("Валентинка создана",
		"valentine_id", valentine.ID,
		"sender_id", senderID,
		"recipient_id", recipientID)

	return valentine, nil
}

// GetSentValentines возвращает отправленные валентинки пользователя
func (uc *ValentineUseCases) GetSentValentines(ctx context.Context, userID int) ([]*domain.Valentine, error) {
	uc.log.Debug("Получение отправленных валентинок", "user_id", userID)

	valentines, err := uc.repo.FindBySender(ctx, userID)
	if err != nil {
		uc.log.Error("Ошибка получения отправленных валентинок",
			"user_id", userID,
			"error", err)
		return nil, err
	}

	uc.log.Debug("Найдено отправленных валентинок",
		"user_id", userID,
		"count", len(valentines))

	return valentines, nil
}

// GetReceivedValentines возвращает полученные валентинки пользователя
func (uc *ValentineUseCases) GetReceivedValentines(ctx context.Context, userID int) ([]*domain.Valentine, error) {
	uc.log.Debug("Получение полученных валентинок", "user_id", userID)

	// Получаем только отправленные валентинки тестово пока что похуй что в конце ставлю true
	valentines, err := uc.repo.FindByRecipient(ctx, userID, true) // ВСЯ ПРОБЛЕМА БЫЛА В false БЛЯ ИДИ НАХУЙ СУКААА
	if err != nil {
		uc.log.Error("Ошибка получения полученных валентинок",
			"user_id", userID,
			"error", err)
		return nil, err
	}
	uc.log.Info("вот скока хуйни валентинок отправленных", valentines)
	// Фильтруем только те, которые можно просматривать
	var viewableValentines []*domain.Valentine
	for _, v := range valentines {
		if v.CanBeViewedByRecipient() {
			viewableValentines = append(viewableValentines, v)
		}
	}

	uc.log.Debug("Найдено полученных валентинок",
		"user_id", userID,
		"total", len(valentines),
		"viewable", len(viewableValentines))

	//	 Помечаем валентинки как открытые
	//	for _, v := range viewableValentines {
	//		if !v.Opened {
	//			if err := uc.repo.MarkAsOpened(ctx, v.ID); err != nil {
	//				uc.log.Warn("Ошибка пометки валентинки как открытой",
	//					"valentine_id", v.ID,
	//					"error", err)
	//			}
	//		}
	//	}

	return valentines, nil
}

// CanViewReceived проверяет, можно ли просматривать полученные валентинки сегодня
func (uc *ValentineUseCases) CanViewReceived() bool {
	if uc.testMode {
		return true
	}
	now := time.Now()
	canView := now.Month() == time.February && now.Day() >= 14

	uc.log.Debug("Проверка возможности просмотра полученных валентинок",
		"current_date", now.Format("2006-01-02"),
		"can_view", canView)

	return canView
}

// GetStats возвращает статистику пользователя
func (uc *ValentineUseCases) GetStats(ctx context.Context, userID int) (int, int, error) {
	return uc.repo.GetStats(ctx, userID)
}

// GetUnsentValentines возвращает неотправленные валентинки
func (uc *ValentineUseCases) GetUnsentValentines(ctx context.Context) ([]*domain.Valentine, error) {
	return uc.repo.FindUnsent(ctx)
}

// MarkValentineAsSent помечает валентинку как отправленную
func (uc *ValentineUseCases) MarkValentineAsSent(ctx context.Context, id string) error {
	return uc.repo.MarkAsSent(ctx, id)
}

// UpdateValentinePhoto обновляет фото валентинки
func (uc *ValentineUseCases) UpdateValentinePhoto(ctx context.Context, id string, photoURL string) error {
	return uc.repo.UpdatePhotoURL(ctx, id, photoURL)
}

// parseRecipient преобразует ссылку в ID и сохраняет оригинал
// Всегда резолвим screen_name через VK API. При неудаче → ошибка.
func (uc *ValentineUseCases) parseRecipient(link string) (recipientID int, recipientOriginal string, err error) {
	link = strings.TrimSpace(link)
	recipientOriginal = link

	// 1. Пробуем как числовой ID
	if id, err := strconv.Atoi(link); err == nil && id > 0 {
		return id, recipientOriginal, nil
	}

	// 2. Извлекаем screen_name из ссылки
	screenName := extractScreenName(link)
	if screenName == "" {
		return 0, recipientOriginal, fmt.Errorf("не удалось извлечь screen_name из ссылки: %s", link)
	}

	// 3. Резолвим screen_name через VK API
	resolved, err := uc.vk.UtilsResolveScreenName(api.Params{
		"screen_name": screenName,
	})
	if err != nil {
		return 0, recipientOriginal, fmt.Errorf("ошибка VK API при резолве %s: %w", screenName, err)
	}
	if resolved.Type != "user" {
		return 0, recipientOriginal, fmt.Errorf("screen_name %s не является пользователем (тип: %s)", screenName, resolved.Type)
	}

	return resolved.ObjectID, recipientOriginal, nil
}

// extractScreenName — вспомогательная функция
func extractScreenName(link string) string {
	link = strings.TrimPrefix(link, "https://")
	link = strings.TrimPrefix(link, "http://")
	link = strings.TrimPrefix(link, "vk.com/")
	link = strings.TrimPrefix(link, "m.vk.com/")
	link = strings.TrimPrefix(link, "@")
	link = strings.TrimPrefix(link, "id") // на случай id123
	// после удалений остаётся либо короткое имя, либо пусто
	if link == "" {
		return ""
	}
	return link
}
