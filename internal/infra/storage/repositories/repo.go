package repositories

import (
	"context"
	"time"

	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/domain"
	"gorm.io/gorm"
)

// Методы взаимодействия с базой данных
type Repository interface {
	Create(ctx context.Context, valentine *domain.Valentine) error
	GetAllSender(ctx context.Context, senderID int) ([]*domain.Valentine, error)
	GetAllReciever(ctx context.Context, recieverID int) ([]*domain.Valentine, error)
	Exist(ctx context.Context, senderID, recieverID, date time.Time) (bool, error)
}

type GORMValentineRepository struct {
	db *gorm.DB
}

// NewGORMValentineRepo создает новый экземпляр репозитория
func NewGORMValentineRepo(db *gorm.DB) *GORMValentineRepository {
	return &GORMValentineRepository{db: db}
}

func (g *GORMValentineRepository) Create(ctx context.Context, valentine *domain.Valentine) error {
	return g.db.WithContext(ctx).Create(valentine).Error
}

func (g *GORMValentineRepository) GetAllSender(ctx context.Context, senderID int) ([]*domain.Valentine, error) {
	var valentines []*domain.Valentine
	err := g.db.WithContext(ctx).
		Where("sender_id = ?", senderID).
		// Order("sent_at DESC").
		Find(&valentines).Error

	return valentines, err
}

func (g *GORMValentineRepository) GetAllReciever(ctx context.Context, recieverID int) ([]*domain.Valentine, error) {
	var valentines []*domain.Valentine
	err := g.db.WithContext(ctx).
		Where("recipient_id = ?", recieverID).
		// Order("sent_at DESC").
		Find(&valentines).Error

	return valentines, err
}

func (g *GORMValentineRepository) Exist(ctx context.Context, senderID, receiverID int, date time.Time) (bool, error) {
	var count int64
	// startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	// endOfDay := startOfDay.Add(24 * time.Hour)

	err := g.db.WithContext(ctx).Model(&domain.Valentine{}).
		// Where("sender_id = ? AND recipient_id = ? AND sent_at >= ? AND sent_at < ?",
		Where("sender_id = ? AND recipient_id = ?",
			senderID, receiverID).
		// senderID, receiverID, startOfDay, endOfDay).
		Count(&count).Error

	return count > 0, err
}
