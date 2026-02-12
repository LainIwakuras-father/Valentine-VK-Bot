package vk

import (
	"bytes"
	//"context"
	"encoding/json"
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

// uploadResponse — ответ от сервера загрузки фото
type uploadResponse struct {
	Server int    `json:"server"`
	Photo  string `json:"photo"`
	Hash   string `json:"hash"`
}

// uploadPhotoToServer загружает байты фото на сервер VK и возвращает ответ
func UploadPhotoToServer(uploadURL string, photoBytes []byte) (*uploadResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("photo", "valentine.jpg")
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, bytes.NewReader(photoBytes))
	if err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result uploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

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
