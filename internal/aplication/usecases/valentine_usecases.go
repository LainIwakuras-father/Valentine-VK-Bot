package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/domain"
	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/infra/storage/repositories"
)

// ValentineUseCases содержит бизнес-логику работы с валентинками
type ValentineUseCases struct {
	repo repositories.GORMValentineRepository
}

// NewValentineUseCases создает новый экземпляр use cases
func NewValentineUseCases(repo repositories.GORMValentineRepository) *ValentineUseCases {
	return &ValentineUseCases{repo: repo}
}

// сценарий сохранения в БД валентинки и уведомления о ней получателя
func (u *ValentineUseCases) SendValentine(
	ctx context.Context,
	senderID int,
	recipientLink string,
	message string,
	isAnonymous bool,
	imageType string,
) (*domain.Valentine, error) {
	// Парсим ID получателя из ссылки
	recipientID, err := u.parseRecipientID(recipientLink)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга получателя: %w", err)
	}

	// Проверяем, что отправитель не отправляет себе
	if senderID == recipientID {
		return nil, fmt.Errorf("нельзя отправить валентинку самому себе")
	}

	// Создаем валентинку
	valentine := &domain.Valentine{
		// ID:          uuid.New().String(),
		SenderID:    senderID,
		RecipientID: recipientID,
		Message:     message,
		// ImageType:   imageType,
		// IsAnonymous: isAnonymous,
		// CreatedAt:   time.Now(),
		// SentAt:      nil, // Будет отправлено 14 февраля
		// Opened:      false,
	}

	// Сохраняем в БД
	if err := u.repo.Create(ctx, valentine); err != nil {
		return nil, fmt.Errorf("ошибка сохранения валентинки: %w", err)
	}

	return valentine, nil
}

// сценарий просмотра полученный сообщений
func (u *ValentineUseCases) GetReceivedValentines(ctx context.Context, userID int) ([]*domain.Valentine, error) {
	// Проверяем дату
	now := time.Now()
	if now.Month() != time.February || now.Day() != 14 {
		return nil, fmt.Errorf("полученные валентинки можно посмотреть только 14 февраля")
	}

	valentines, err := u.repo.GetAllReciever(ctx, userID)
	if err != nil {
		return nil, err
	}

	return valentines, nil
}

// сценарий просмотра отправленных сообщений
// GetSentValentines возвращает отправленные валентинки пользователя
func (u *ValentineUseCases) GetSentValentines(ctx context.Context, userID int) ([]*domain.Valentine, error) {
	return u.repo.GetAllSender(ctx, userID)
}

// вспомогательная функция парсинга
func (u *ValentineUseCases) parseRecipientID(link string) (int, error) {
	// Убираем пробелы
	link = strings.TrimSpace(link)

	// Если это уже число
	var recipientID int
	_, err := fmt.Sscanf(link, "%d", &recipientID)
	if err == nil && recipientID > 0 {
		return recipientID, nil
	}

	// Пробуем извлечь из ссылки vk.com
	// Примеры: https://vk.com/id123456, vk.com/id123456, @id123456

	// Ищем "id" в ссылке
	link = strings.ToLower(link)

	// Пробуем разные форматы
	patterns := []string{
		"vk.com/id",
		"vk.com/",
		"id",
		"@",
	}

	for _, pattern := range patterns {
		if idx := strings.LastIndex(link, pattern); idx != -1 {
			idStr := link[idx+len(pattern):]
			// Убираем все нецифры
			idStr = strings.Map(func(r rune) rune {
				if r >= '0' && r <= '9' {
					return r
				}
				return -1
			}, idStr)

			if len(idStr) > 0 {
				fmt.Sscanf(idStr, "%d", &recipientID)
				if recipientID > 0 {
					return recipientID, nil
				}
			}
		}
	}

	return 0, fmt.Errorf("не удалось распознать ID получателя")
}
