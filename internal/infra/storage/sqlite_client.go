package storage

import (
	"fmt"
	"os"

	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewSqliteDB() (*gorm.DB, error) {
	// Берём путь из переменной окружения, иначе стандартный /app/data/valentine.db
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		// Проверим, существует ли папка /app/data — если нет, значит локальный запуск
		dbPath = "./valentine.db"
	}
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("Ошибка подключения БД %w", err)
	}

	// Автоматическая миграция
	if err := db.AutoMigrate(&domain.Valentine{}); err != nil {
		return nil, fmt.Errorf("ошибка миграции БД: %w", err)
	}
	return db, nil
}

// CloseDB закрывает соединение с базой данных
func CloseDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
