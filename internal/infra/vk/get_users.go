package vk

import (
	"log"

	"github.com/SevereCloud/vksdk/v3/api"
)

func GetScreenName(vk *api.VK, userID int) string {
	users, err := vk.UsersGet(api.Params{
		"user_ids": userID, // ID пользователя (число или строка)
		"fields":   "screen_name",
	})
	if err != nil {
		log.Fatal(err)
	}
	if len(users) > 0 {
		screenName := users[0].ScreenName // короткое имя без @
		return screenName
	}
	return ""
}
