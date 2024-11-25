package service

import (
	"context"
	"github.com/gorilla/websocket"
	"tgbot"
	"tgbot/internal/repository"
)

type Bot interface {
}
type User interface {
	SignIn(string, string) (tgbot.User, error)
}
type Pixel interface {
	DrawRectangel(x0, y0, x1, y1 int, color string, id int)
	SendPixel(pixels []tgbot.PixelClick, conn *websocket.Conn)
}
type Converter interface {
	Download(string, string) error
	Pixelize(string, string, int, int, chan *tgbot.Pixel, context.Context, int) ([]byte, error)
	SendPixels(string, string, int, int, tgbot.Point, int) error
}
type Service struct {
	Bot
	User
	Pixel
	Converter
}

func NewService(repo *repository.Repository) *Service {
	return &Service{
		Bot:       NewBotService(repo.Bot),
		User:      NewUserService(repo.User),
		Pixel:     NewPixelService(),
		Converter: NewConverterService(NewPixelService()),
	}
}
