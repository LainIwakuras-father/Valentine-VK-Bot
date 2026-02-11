package domain

import (
	"strconv"
	"time"

	"gorm.io/gorm"
)

type Valentine struct {
	gorm.Model
	SenderID    int    `gorm:"index;not null" json:"sender_id"`
	RecipientID int    `gorm:"index;not null" json:"recipient_id"`
	Message     string `gorm:"type:text;not null" json:"message"`
	ImageType   string `gorm:"size:20"`
	ImageID     string `gorm:"size:100"`
	IsAnonymous bool   `gorm:"default:false"`
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
