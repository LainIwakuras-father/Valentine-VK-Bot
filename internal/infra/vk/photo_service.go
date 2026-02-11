package vk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/SevereCloud/vksdk/v3/api"
	//"github.com/SevereCloud/vksdk/v3/object"
)

// PhotoService сервис для работы с фото VK
type PhotoService struct {
	vk  *api.VK
	log *slog.Logger
}

// NewPhotoService создает новый сервис для работы с фото
func NewPhotoService(vk *api.VK, log *slog.Logger) *PhotoService {
	return &PhotoService{
		vk:  vk,
		log: log.With("component", "photo_service"),
	}
}

// UploadPhoto загружает фото на сервер VK
func (s *PhotoService) UploadPhoto(ctx context.Context, userID int, photoBytes []byte) (string, error) {
	s.log.Debug("Начало загрузки фото", "user_id", userID, "size_bytes", len(photoBytes))

	// 1. Получаем адрес сервера для загрузки
	uploadServer, err := s.vk.PhotosGetMessagesUploadServer(api.Params{
		"peer_id": userID,
	})
	if err != nil {
		s.log.Error("Ошибка получения сервера загрузки", "error", err)
		return "", fmt.Errorf("ошибка получения сервера загрузки: %w", err)
	}

	// 2. Загружаем фото на полученный сервер
	photoURL, err := s.uploadToServer(uploadServer.UploadURL, photoBytes)
	if err != nil {
		s.log.Error("Ошибка загрузки фото на сервер", "error", err)
		return "", fmt.Errorf("ошибка загрузки фото: %w", err)
	}

	// 3. Сохраняем фото в VK
	photos, err := s.vk.PhotosSaveMessagesPhoto(api.Params{
		"photo":  photoURL,
		"server": photoURL,
		"hash":   "hash", // В реальном коде нужно получить из ответа сервера
	})
	if err != nil {
		s.log.Error("Ошибка сохранения фото в VK", "error", err)
		return "", fmt.Errorf("ошибка сохранения фото: %w", err)
	}

	if len(photos) == 0 {
		s.log.Error("Фото не было сохранено")
		return "", fmt.Errorf("фото не было сохранено")
	}

	// 4. Формируем ссылку на фото
	photo := photos[0]
	photoAttachment := fmt.Sprintf("photo%d_%d", photo.OwnerID, photo.ID)

	s.log.Info("Фото успешно загружено",
		"photo_id", photo.ID,
		"owner_id", photo.OwnerID,
		"attachment", photoAttachment)

	return photoAttachment, nil
}

// uploadToServer загружает фото на указанный сервер
func (s *PhotoService) uploadToServer(uploadURL string, photoBytes []byte) (string, error) {
	// Создаем multipart запрос
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Добавляем файл
	part, err := writer.CreateFormFile("photo", "valentine.jpg")
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(part, bytes.NewReader(photoBytes)); err != nil {
		return "", err
	}

	writer.Close()

	// Отправляем запрос
	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Читаем ответ
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// В реальном коде нужно парсить JSON ответ
	// Для демо просто возвращаем строку
	return string(respBody), nil
}

// SendPhotoMessage отправляет сообщение с фото
func SendPhotoMessage(vk *api.VK, userID int, message string, photoAttachment string) error {
	params := api.Params{
		"user_id":    userID,
		"message":    message,
		"attachment": photoAttachment,
		"random_id":  0,
	}

	_, err := vk.MessagesSend(params)
	return err
}

// GetPhotoAttachment создает attachment для фото
func GetPhotoAttachment(ownerID, photoID int) string {
	return fmt.Sprintf("photo%d_%d", ownerID, photoID)
}

// IsValidPhotoURL проверяет, является ли строка допустимым URL фото
func IsValidPhotoURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Проверяем, что это HTTP/HTTPS
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	// Проверяем расширение файла
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	for _, ext := range validExtensions {
		if strings.HasSuffix(strings.ToLower(u.Path), ext) {
			return true
		}
	}

	return false
}
