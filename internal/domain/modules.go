package domain

import (
	"strconv"
	"time"
	//"gorm.io/gorm"
)

type Valentine struct {
	ID          string `gorm:"primaryKey"`
	SenderID    int    `gorm:"index;not null" json:"sender_id"`
	RecipientID int    `gorm:"index;not null" json:"recipient_id"`
	Message     string `gorm:"type:text;not null" json:"message"`
	ImageType   string `gorm:"size:20"`
	ImageID     string `gorm:"size:100"`
	PhotoURL    string `gorm:"size:500"` // Ссылка на загруженное фото
	IsAnonymous bool   `gorm:"default:false"`
	SentAt      *time.Time
	// Когда была отправлена (nil = еще не отправлена)
	Opened bool `gorm:"default:false"`
}

// TableName задает имя таблицы в БД
func (Valentine) TableName() string {
	return "valentines"
}

// CanViewReceived проверяет, можно ли просматривать полученные валентинки
// (только 14 февраля, как указано в требованиях)
func (v *Valentine) CanViewReceived(now time.Time) bool {
	return now.Month() == time.February && now.Day() == 14
}

// GetSenderDisplay возвращает отображаемое имя отправителя
func (v *Valentine) GetSenderDisplay() string {
	if v.IsAnonymous {
		return "Аноним"
	}
	return "ID" + strconv.Itoa(v.SenderID)
}

// FormatMessage форматирует сообщение для отображения
func (v *Valentine) FormatMessage() string {
	message := v.Message
	if len(message) > 100 {
		message = message[:100] + "..."
	}
	return message
}

// IsSent проверяет, отправлена ли валентинка
func (v *Valentine) IsSent() bool {
	return v.SentAt != nil
}

// CanBeViewedByRecipient проверяет, может ли получатель просмотреть валентинку
func (v *Valentine) CanBeViewedByRecipient() bool {
	if !v.IsSent() {
		return false // Не отправлена
	}

	// Можно просматривать, если отправлена 14 февраля или позже
	sentDate := v.SentAt
	if sentDate == nil {
		return false
	}

	// Если сегодня 14 февраля или позже, чем дата отправки
	now := time.Now()

	// Проверяем, что валентинка отправлена в этом году и можно просматривать после 14 февраля
	if sentDate.Year() == now.Year() {
		// Можно просматривать с 14 февраля
		viewingStart := time.Date(sentDate.Year(), time.February, 14, 0, 0, 0, 0, sentDate.Location())
		return now.After(viewingStart) || now.Equal(viewingStart)
	}

	return true // Если отправлена в прошлом году, можно смотреть
}
