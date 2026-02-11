package domain

import (
	"time"

	"gorm.io/gorm"
)

type Valentine struct {
	gorm.Model
	SenderID    int    `gorm:"index;not null" json:"sender_id"`
	RecipientID int    `gorm:"index;not null" json:"recipient_id"`
	Message     string `gorm:"type:text;not null" json:"message"`
	// ImageType   string
	// ImageID     int
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
