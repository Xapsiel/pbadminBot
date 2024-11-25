package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func isValidHexColor(color string) bool {
	if len(color) != 7 {
		return false
	}
	if color[0] != '#' {
		return false
	}
	for i := 1; i < 7; i++ {
		if !((color[i] >= '0' && color[i] <= '9') || (color[i] >= 'a' && color[i] <= 'f') || (color[i] >= 'A' && color[i] <= 'F')) {
			return false
		}
	}
	return true
}

// Добавляем клавиатуру для выбора действия
var actionNumber = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Закрасить область"),
		tgbotapi.NewKeyboardButton("Нарисовать пиксельарт"),
	),
)

// Добавляем клавиатуру для подтверждения или отмены действия
var Review = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Да"),
		tgbotapi.NewKeyboardButton("Нет"),
	),
)
