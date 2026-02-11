package repositories

import (
	"context"
	"time"

	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/domain"
	"gorm.io/gorm"
)

// ValentineRepository интерфейс репозитория для работы с валентинками
type ValentineRepository interface {
	Create(ctx context.Context, valentine *domain.Valentine) error
	FindByID(ctx context.Context, id string) (*domain.Valentine, error)
	FindBySender(ctx context.Context, senderID int) ([]*domain.Valentine, error)
	FindByRecipient(ctx context.Context, recipientID int, includeUnsent bool) ([]*domain.Valentine, error)
	FindUnsent(ctx context.Context) ([]*domain.Valentine, error)
	MarkAsOpened(ctx context.Context, id string) error
	MarkAsSent(ctx context.Context, id string) error
	GetStats(ctx context.Context, userID int) (int, int, error)
	UpdatePhotoURL(ctx context.Context, id string, photoURL string) error
}

// GORMValentineRepository реализация репозитория на GORM
type GORMValentineRepository struct {
	db *gorm.DB
}

// NewGORMValentineRepo создает новый репозиторий
func NewGORMValentineRepo(db *gorm.DB) ValentineRepository {
	return &GORMValentineRepository{db: db}
}

// Create создает новую валентинку
func (r *GORMValentineRepository) Create(ctx context.Context, valentine *domain.Valentine) error {
	return r.db.WithContext(ctx).Create(valentine).Error
}

// FindByID находит валентинку по ID
func (r *GORMValentineRepository) FindByID(ctx context.Context, id string) (*domain.Valentine, error) {
	var valentine domain.Valentine
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&valentine).Error
	if err != nil {
		return nil, err
	}
	return &valentine, nil
}

// FindBySender находит валентинки по отправителю
func (r *GORMValentineRepository) FindBySender(ctx context.Context, senderID int) ([]*domain.Valentine, error) {
	var valentines []*domain.Valentine
	err := r.db.WithContext(ctx).
		Where("sender_id = ?", senderID).
		// Order("created_at DESC").
		Find(&valentines).Error
	return valentines, err
}

// FindByRecipient находит валентинки по получателю
// includeUnsent - включать ли неотправленные валентинки
func (r *GORMValentineRepository) FindByRecipient(ctx context.Context, recipientID int, includeUnsent bool) ([]*domain.Valentine, error) {
	var valentines []*domain.Valentine
	query := r.db.WithContext(ctx).
		Where("recipient_id = ?", recipientID)

	if !includeUnsent {
		query = query.Where("sent_at IS NOT NULL")
	}
	// Order("created_at DESC")
	err := query.Find(&valentines).Error
	return valentines, err
}

// FindUnsent находит все неотправленные валентинки
func (r *GORMValentineRepository) FindUnsent(ctx context.Context) ([]*domain.Valentine, error) {
	var valentines []*domain.Valentine
	err := r.db.WithContext(ctx).
		Where("sent_at IS NULL").
		Find(&valentines).Error
	return valentines, err
}

// MarkAsOpened помечает валентинку как открытую
func (r *GORMValentineRepository) MarkAsOpened(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&domain.Valentine{}).
		Where("id = ?", id).
		Update("opened", true).Error
}

// MarkAsSent помечает валентинку как отправленную
func (r *GORMValentineRepository) MarkAsSent(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.Valentine{}).
		Where("id = ?", id).
		Update("sent_at", now).Error
}

// GetStats возвращает статистику пользователя
func (r *GORMValentineRepository) GetStats(ctx context.Context, userID int) (int, int, error) {
	var sentCount int64
	err := r.db.WithContext(ctx).
		Model(&domain.Valentine{}).
		Where("sender_id = ?", userID).
		Count(&sentCount).Error
	if err != nil {
		return 0, 0, err
	}

	var receivedCount int64
	err = r.db.WithContext(ctx).
		Model(&domain.Valentine{}).
		Where("recipient_id = ? AND sent_at IS NOT NULL", userID).
		Count(&receivedCount).Error

	return int(sentCount), int(receivedCount), err
}

// UpdatePhotoURL обновляет URL фото валентинки
func (r *GORMValentineRepository) UpdatePhotoURL(ctx context.Context, id string, photoURL string) error {
	return r.db.WithContext(ctx).
		Model(&domain.Valentine{}).
		Where("id = ?", id).
		Update("photo_url", photoURL).Error
}
