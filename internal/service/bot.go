package service

import (
	"tgbot/internal/repository"
)

type BotService struct {
	repo repository.Bot
}

func NewBotService(repo repository.Bot) *BotService {
	return &BotService{
		repo: repo,
	}
}

//func (b *BotService) SignIn(login, password string) error {
//	login = strings.Replace(login, " ", "", -1)
//	password = strings.Replace(password, " ", "", -1)
//}
