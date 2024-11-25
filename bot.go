package tgbot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Bot struct {
	BotObj *tgbotapi.BotAPI `json:"bot_obj"`
}
