package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/domain"
	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/infra/storage/repositories"
	"github.com/google/uuid"
)

// ValentineUseCases бизнес-логика для работы с валентинками
type ValentineUseCases struct {
	repo repositories.ValentineRepository
	log  *slog.Logger
}

// NewValentineUseCases создает новый use case
func NewValentineUseCases(repo repositories.ValentineRepository, log *slog.Logger) *ValentineUseCases {
	return &ValentineUseCases{
		repo: repo,
		log:  log.With("component", "valentine_usecases"),
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
	recipientID, err := uc.parseRecipientID(recipientLink)
	if err != nil {
		uc.log.Error("Ошибка парсинга получателя", "error", err)
		return nil, fmt.Errorf("ошибка парсинга получателя: %w", err)
	}

	// Проверяем, что отправитель не отправляет себе
	if senderID == recipientID {
		err := fmt.Errorf("нельзя отправить валентинку самому себе")
		uc.log.Error("Попытка отправить валентинку себе", "error", err)
		return nil, err
	}

	// Создаем валентинку
	valentine := &domain.Valentine{
		ID:          uuid.New().String(),
		SenderID:    senderID,
		RecipientID: recipientID,
		Message:     message,
		ImageType:   imageType,
		PhotoURL:    photoURL,
		IsAnonymous: isAnonymous,
		Opened:      false,
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

	// Получаем только отправленные валентинки
	valentines, err := uc.repo.FindByRecipient(ctx, userID, false)
	if err != nil {
		uc.log.Error("Ошибка получения полученных валентинок",
			"user_id", userID,
			"error", err)
		return nil, err
	}

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
	for _, v := range viewableValentines {
		if !v.Opened {
			if err := uc.repo.MarkAsOpened(ctx, v.ID); err != nil {
				uc.log.Warn("Ошибка пометки валентинки как открытой",
					"valentine_id", v.ID,
					"error", err)
			}
		}
	}

	return viewableValentines, nil
}

// CanViewReceived проверяет, можно ли просматривать полученные валентинки сегодня
func (uc *ValentineUseCases) CanViewReceived() bool {
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

func (uc *ValentineUseCases) parseRecipientID(link string) (int, error) {
	link = strings.TrimSpace(link)
	// Удаляем протокол и домен
	link = strings.TrimPrefix(link, "https://")
	link = strings.TrimPrefix(link, "http://")
	link = strings.TrimPrefix(link, "vk.com/")
	link = strings.TrimPrefix(link, "m.vk.com/")
	link = strings.TrimPrefix(link, "@")
	link = strings.TrimPrefix(link, "id")
	link = strings.TrimPrefix(link, "club") // для групп, но нам нужны пользователи

	// Теперь link должен содержать только цифры
	var id int
	_, err := fmt.Sscanf(link, "%d", &id)
	if err != nil || id <= 0 {
		uc.log.Warn("Не удалось распарсить ID", "input", link)
		// Возвращаем 0, чтобы вызвать ошибку
		return 0, fmt.Errorf("некорректный ID получателя: %s", link)
	}
	return id, nil
}
