package vk

import (
	"log"

	"github.com/SevereCloud/vksdk/v3/api"
	"github.com/SevereCloud/vksdk/v3/object"
)

// SendMessage отправляет простое сообщение
func SendMessage(vk *api.VK, userID int, message string) error {
	return SendKeyboard(vk, userID, message, nil)
}

func SendKeyboard(vk *api.VK, userID int, message string, keyboard *object.MessagesKeyboard) error {
	params := api.Params{
		"peer_id":   userID,
		"message":   message,
		"random_id": 0,
	}

	if keyboard != nil {
		params["keyboard"] = keyboard
	}

	_, err := vk.MessagesSend(params)
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
	return err
}

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
					Color: "positive",
				},
			},
			{
				{
					Action: object.MessagesKeyboardButtonAction{
						Type:  "text",
						Label: "📤 Мои отправленные",
					},
					Color: "primary",
				},
				{
					Action: object.MessagesKeyboardButtonAction{
						Type:  "text",
						Label: "📥 Мои полученные",
					},
					Color: "primary",
				},
			},
		},
	}
}

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
					Color: "primary",
				},
			},
			{
				{
					Action: object.MessagesKeyboardButtonAction{
						Type:  "text",
						Label: "Нет",
					},
					Color: "secondary",
				},
			},
		},
	}
}

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

func NewPremadeImagesKeyboard(images []map[string]string) *object.MessagesKeyboard {
	buttons := make([][]object.MessagesKeyboardButton, 0)

	// Группируем по 2 кнопки в ряд
	for i := 0; i < len(images); i += 2 {
		row := make([]object.MessagesKeyboardButton, 0)

		// Первая кнопка в ряду
		row = append(row, object.MessagesKeyboardButton{
			Action: object.MessagesKeyboardButtonAction{
				Type:  "text",
				Label: images[i]["description"],
			},
			Color: "secondary",
		})

		// Вторая кнопка в ряду (если есть)
		if i+1 < len(images) {
			row = append(row, object.MessagesKeyboardButton{
				Action: object.MessagesKeyboardButtonAction{
					Type:  "text",
					Label: images[i+1]["description"],
				},
				Color: "secondary",
			})
		}

		buttons = append(buttons, row)
	}

	// Добавляем кнопку "Назад"
	buttons = append(buttons, []object.MessagesKeyboardButton{
		{
			Action: object.MessagesKeyboardButtonAction{
				Type:  "text",
				Label: "« Назад",
			},
			Color: "negative",
		},
	})

	return &object.MessagesKeyboard{
		OneTime: true,
		Buttons: buttons,
	}
}
