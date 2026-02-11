package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/LainIwakuras-father/Valentine-VK-Bot/internal/domain"
	//	vkkeyboard "github.com/LainIwakuras-father/Valentine-VK-Bot/internal/infra/vk"
	"github.com/SevereCloud/vksdk/v3/api"
)

// Scheduler планировщик для отправки валентинок
type Scheduler struct {
	vk               *api.VK
	valentineService *ValentineUseCases
	log              *slog.Logger
	stopChan         chan bool
}

// NewScheduler создает новый планировщик
func NewScheduler(vk *api.VK, service *ValentineUseCases, log *slog.Logger) *Scheduler {
	return &Scheduler{
		vk:               vk,
		valentineService: service,
		log:              log.With("component", "scheduler"),
		stopChan:         make(chan bool),
	}
}

// Start запускает планировщик
func (s *Scheduler) Start() {
	s.log.Info("Запуск планировщика валентинок")

	go func() {
		ticker := time.NewTicker(1 * time.Hour) // Проверяем каждый час
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.checkAndSendValentines(context.Background())
			case <-s.stopChan:
				s.log.Info("Планировщик остановлен")
				return
			}
		}
	}()
}

// Stop останавливает планировщик
func (s *Scheduler) Stop() {
	s.stopChan <- true
}

// checkAndSendValentines проверяет и отправляет валентинки
func (s *Scheduler) checkAndSendValentines(ctx context.Context) {
	now := time.Now()

	// Проверяем, сегодня ли 14 февраля
	if now.Month() != time.February || now.Day() != 14 {
		s.log.Debug("Сегодня не 14 февраля, пропускаем отправку",
			"current_date", now.Format("2006-01-02"))
		return
	}

	s.log.Info("14 февраля! Проверяем неотправленные валентинки")

	// Получаем неотправленные валентинки
	valentines, err := s.valentineService.GetUnsentValentines(ctx)
	if err != nil {
		s.log.Error("Ошибка получения неотправленных валентинок", "error", err)
		return
	}

	if len(valentines) == 0 {
		s.log.Info("Нет неотправленных валентинок")
		return
	}

	s.log.Info("Найдены неотправленные валентинки", "count", len(valentines))

	// Отправляем каждую валентинку
	for _, valentine := range valentines {
		if err := s.sendValentine(ctx, valentine); err != nil {
			s.log.Error("Ошибка отправки валентинки",
				"valentine_id", valentine.ID,
				"error", err)
			continue
		}

		// Помечаем как отправленную
		if err := s.valentineService.MarkValentineAsSent(ctx, valentine.ID); err != nil {
			s.log.Error("Ошибка пометки валентинки как отправленной",
				"valentine_id", valentine.ID,
				"error", err)
		}

		s.log.Info("Валентинка отправлена",
			"valentine_id", valentine.ID,
			"sender_id", valentine.SenderID,
			"recipient_id", valentine.RecipientID)
	}
}

// sendValentine отправляет валентинку получателю
func (s *Scheduler) sendValentine(ctx context.Context, valentine *domain.Valentine) error {
	message := "💌 Вы получили валентинку!\n\n"

	if valentine.IsAnonymous {
		message += "🎭 От: Аноним\n"
	} else {
		message += fmt.Sprintf("👤 От: ID%d\n", valentine.SenderID)
	}

	message += "💌 Сообщение: " + valentine.Message + "\n\n"
	message += "💖 С Днем Святого Валентина!"

	// Если есть фото, отправляем с фото
	if valentine.PhotoURL != "" {
		message += "\n\n📷 К валентинке приложено фото!"
		// В реальном проекте нужно было бы прикрепить фото
	}

	// Отправляем сообщение через VK API
	_, err := s.vk.MessagesSend(api.Params{
		"user_id":   valentine.RecipientID,
		"message":   message,
		"random_id": 0,
	})

	return err
}
