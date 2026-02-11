package config

import (
	"log/slog"
	"os"
)

// SetupLogger настраивает структурированный логгер
func SetupLogger() *slog.Logger {
	// Настраиваем JSON логгер для продакшена
	// или текстовый для разработки
	logLevel := slog.LevelInfo

	if os.Getenv("DEBUG") == "true" {
		logLevel = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				// Форматируем время для лучшей читаемости
				return slog.Attr{}
			}
			return a
		},
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

// WithRequest добавляет ID запроса к логам
func WithRequest(logger *slog.Logger, requestID string) *slog.Logger {
	return logger.With("request_id", requestID)
}

// WithUser добавляет ID пользователя к логам
func WithUser(logger *slog.Logger, userID int) *slog.Logger {
	return logger.With("user_id", userID)
}
