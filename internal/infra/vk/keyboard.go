package vk

import (
	"github.com/SevereCloud/vksdk/v3/api"
	"github.com/SevereCloud/vksdk/v3/object"
)

// SendKeyboard отправляет сообщение с клавиатурой
func SendKeyboard(vk *api.VK, userID int, message string, keyboard *object.MessagesKeyboard) error {
	params := api.Params{
		"user_id":   userID,
		"message":   message,
		"random_id": 0,
	}

	if keyboard != nil {
		params["keyboard"] = keyboard
	}

	_, err := vk.MessagesSend(params)
	return err
}

// SendMessage отправляет простое сообщение без клавиатуры
func SendMessage(vk *api.VK, userID int, message string) error {
	return SendKeyboard(vk, userID, message, nil)
}

// NewStartKeyboard создает клавиатуру главного меню
func NewStartKeyboard() *object.MessagesKeyboard {
	return &object.MessagesKeyboard{
		OneTime: false,
		Buttons: [][]object.MessagesKeyboardButton{
			{
				{
					Action: object.MessagesKeyboardButtonAction{
						Type:  "text",
						Label: "💌 Отправить валентинку",
					},
					Color: "primary",
				},
			},
			{
				{
					Action: object.MessagesKeyboardButtonAction{
						Type:  "text",
						Label: "📤 Мои отправленные",
					},
					Color: "secondary",
				},
				{
					Action: object.MessagesKeyboardButtonAction{
						Type:  "text",
						Label: "📥 Мои полученные",
					},
					Color: "secondary",
				},
			},
		},
	}
}

// NewAnonymityKeyboard создает клавиатуру для выбора анонимности
func NewAnonymityKeyboard() *object.MessagesKeyboard {
	return &object.MessagesKeyboard{
		OneTime: true,
		Buttons: [][]object.MessagesKeyboardButton{
			{
				{
					Action: object.MessagesKeyboardButtonAction{
						Type:  "text",
						Label: "Да",
					},
					Color: "positive",
				},
				{
					Action: object.MessagesKeyboardButtonAction{
						Type:  "text",
						Label: "Нет",
					},
					Color: "negative",
				},
			},
		},
	}
}

// NewValentineTypeKeyboard создает клавиатуру для выбора типа валентинки (без отдельного фото)
func NewValentineTypeKeyboard() *object.MessagesKeyboard {
	return &object.MessagesKeyboard{
		OneTime: true,
		Buttons: [][]object.MessagesKeyboardButton{
			{
				{
					Action: object.MessagesKeyboardButtonAction{
						Type:  "text",
						Label: "Заготовленная",
					},
					Color: "positive",
				},
			},
			{
				{
					Action: object.MessagesKeyboardButtonAction{
						Type:  "text",
						Label: "Собственная",
					},
					Color: "primary",
				},
			},
		},
	}
}

// NewTemplateKeyboard создаёт клавиатуру с готовыми валентинками (фото)
func NewTemplateKeyboard() *object.MessagesKeyboard {
	return &object.MessagesKeyboard{
		OneTime: true,
		Inline:  false, // обычная, не инлайн (лучше для ботов)
		Buttons: [][]object.MessagesKeyboardButton{
			{
				{Action: object.MessagesKeyboardButtonAction{Type: "text", Label: "💝 1"}, Color: "primary"},
				{Action: object.MessagesKeyboardButtonAction{Type: "text", Label: "💘 2"}, Color: "primary"},
			},
			{
				{Action: object.MessagesKeyboardButtonAction{Type: "text", Label: "💖 3"}, Color: "primary"},
				{Action: object.MessagesKeyboardButtonAction{Type: "text", Label: "💗 4"}, Color: "primary"},
			},
		},
	}
}

// NewPhotoAfterTextKeyboard — клавиатура для добавления фото после ввода текста
func NewPhotoAfterTextKeyboard() *object.MessagesKeyboard {
	return &object.MessagesKeyboard{
		OneTime: true,
		Buttons: [][]object.MessagesKeyboardButton{
			{
				{
					Action: object.MessagesKeyboardButtonAction{
						Type:  "text",
						Label: "📷 Добавить фото",
					},
					Color: "primary",
				},
				{
					Action: object.MessagesKeyboardButtonAction{
						Type:  "text",
						Label: "⏭ Отправить без фото",
					},
					Color: "secondary",
				},
			},
		},
	}
}

// NewCancelKeyboard создает клавиатуру с кнопкой отмены
func NewCancelKeyboard() *object.MessagesKeyboard {
	return &object.MessagesKeyboard{
		OneTime: true,
		Buttons: [][]object.MessagesKeyboardButton{
			{
				{
					Action: object.MessagesKeyboardButtonAction{
						Type:  "text",
						Label: "❌ Отмена",
					},
					Color: "negative",
				},
			},
		},
	}
}
