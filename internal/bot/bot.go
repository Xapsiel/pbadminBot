package bot

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"tgbot"
	"tgbot/internal/service"
)

const (
	canvasHeigth = 100
	canvasWidth  = 100
	nonlogin     = iota
	login
	active
	paintStart    // стадия получения начальной точки прямоугольника
	paintEnd      // стадия получения конечной точки прямоугольника
	pixelArtStart // стадия ввода верхней левой координаты пиксельарта
	pixelArt
	review
	chooseColor // новая стадия выбора цвета
	chooseWidth
	choseeHeight
	channelBuffer = 40
)

type Bot struct {
	token   string
	service *service.Service
	Users   map[int64]*UserSettings `json:"current_user"`
}

type UserSettings struct {
	ID             int
	FillArea       []tgbot.Point
	DrawPixelArt   tgbot.Point
	PixelArtHeight int
	PixelArtWidth  int
	Stage          int
	Color          string
}

func New(token string, service *service.Service) *Bot {
	return &Bot{service: service, token: token, Users: make(map[int64]*UserSettings)}
}

func (b *Bot) Start() {
	bot, err := tgbotapi.NewBotAPI(b.token)
	if err != nil {
		logrus.Fatal(err)
	}
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatid := update.Message.Chat.ID
		if b.Users[chatid] == nil {
			b.Users[chatid] = &UserSettings{FillArea: make([]tgbot.Point, 0)}
		}
		msg := tgbotapi.NewMessage(chatid, "")

		// Обработка команд и авторизация
		if update.Message.IsCommand() || b.Users[chatid].Stage == login {
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "login":
					msg.Text = "Напишите свой логин и пароль в формате \"ЛОГИН:ПАРОЛЬ\""
					b.Users[chatid].Stage = login
					msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
					bot.Send(msg)
					continue
				}

			} else if b.Users[chatid].Stage == login {
				userMessage := strings.Split(update.Message.Text, ":")
				if len(userMessage) != 2 {
					msg.Text = "Неверный формат. Используйте \"ЛОГИН:ПАРОЛЬ\"."
					bot.Send(msg)
					continue
				}

				login, password := userMessage[0], userMessage[1]
				user, err := b.service.SignIn(login, password)
				if err != nil {
					logrus.Error(err)
					msg.Text = "Ошибка авторизации."
					bot.Send(msg)
					continue
				}
				if user.Permissions != 1 {
					msg.Text = "Ты не админ."
					bot.Send(msg)
					continue
				}
				msg.Text = "Авторизация успешна."
				b.Users[chatid] = &UserSettings{
					ID:    user.ID,
					Stage: active,
				}
				msg.ReplyMarkup = actionNumber
				bot.Send(msg)
				continue
			}
		} else if b.Users[chatid].Stage == active {
			switch update.Message.Text {
			case "Закрасить область":
				b.Users[chatid].Stage = chooseColor
				msg.Text = "Введите цвет для закрашивания области в формате HEX (например, #FF0000 для красного)."
				bot.Send(msg)
				continue
			case "Нарисовать пиксельарт":
				b.Users[chatid].Stage = pixelArtStart
				msg.Text = "Введите координаты верхнего левого угла пиксельарта в формате \"x,y\""
				bot.Send(msg)
				continue
			}
		} else if b.Users[chatid].Stage == chooseColor {
			color := update.Message.Text
			if !isValidHexColor(color) {
				msg.Text = "Неверный формат цвета. Используйте формат HEX (например, #FF0000)."
				bot.Send(msg)
				continue
			}

			b.Users[chatid].Color = color
			b.Users[chatid].Stage = paintStart
			msg.Text = "Введите координаты левого верхнего угла прямоугольника в формате \"x,y\""
			bot.Send(msg)
			continue
		} else if b.Users[chatid].Stage == paintStart {
			startCoord := strings.Split(update.Message.Text, ",")
			if len(startCoord) != 2 {
				msg.Text = "Неверный формат координат."
				bot.Send(msg)
				continue
			}

			x0, err := strconv.Atoi(strings.TrimSpace(startCoord[0]))
			y0, err := strconv.Atoi(strings.TrimSpace(startCoord[1]))
			if err != nil {
				logrus.Error(err)
				msg.Text = "Ошибка конвертации координат."
				bot.Send(msg)
				continue
			}

			b.Users[chatid].FillArea = []tgbot.Point{{X: x0, Y: y0}}
			b.Users[chatid].Stage = paintEnd
			msg.Text = "Введите координаты правого нижнего угла прямоугольника \"x,y\""
			bot.Send(msg)
			continue
		} else if b.Users[chatid].Stage == paintEnd {
			endCoord := strings.Split(update.Message.Text, ",")
			if len(endCoord) != 2 {
				msg.Text = "Неверный формат координат."
				bot.Send(msg)
				continue
			}

			x1, err := strconv.Atoi(strings.TrimSpace(endCoord[0]))
			y1, err := strconv.Atoi(strings.TrimSpace(endCoord[1]))
			if err != nil {
				logrus.Error(err)
				msg.Text = "Ошибка конвертации координат."
				bot.Send(msg)
				continue
			}

			start := b.Users[chatid].FillArea[0]
			b.service.Pixel.DrawRectangel(start.X, start.Y, x1, y1, b.Users[chatid].Color, b.Users[chatid].ID)
			b.Users[chatid].Stage = active
			msg.Text = "Прямоугольник закрашен."
			bot.Send(msg)
			continue
		} else if b.Users[chatid].Stage == pixelArtStart {
			startCoord := strings.Split(update.Message.Text, ",")
			if len(startCoord) != 2 {
				msg.Text = "Неверный формат координат."
				bot.Send(msg)
				continue
			}

			x, err := strconv.Atoi(strings.TrimSpace(startCoord[0]))
			y, err := strconv.Atoi(strings.TrimSpace(startCoord[1]))
			if err != nil {
				logrus.Error(err)
				msg.Text = "Ошибка конвертации координат."
				bot.Send(msg)
				continue
			}

			b.Users[chatid].DrawPixelArt = tgbot.Point{X: x, Y: y}
			b.Users[chatid].Stage = chooseWidth
			msg.Text = fmt.Sprintf("Выберите ширину арта(в пикселях, от 8 до %d)", canvasWidth-b.Users[chatid].DrawPixelArt.X)
			//msg.Text = "Загрузите изображение формата или PNG"
			bot.Send(msg)
			continue
		} else if b.Users[chatid].Stage == chooseWidth {
			width, err := strconv.Atoi(update.Message.Text)
			if err != nil {
				msg.Text = "Введите целое число"
				bot.Send(msg)
				continue
			}
			if width > canvasWidth-b.Users[chatid].DrawPixelArt.X || width < 8 {
				msg.Text = fmt.Sprintf("Введите число от 8 до %d)", canvasWidth-b.Users[chatid].DrawPixelArt.Y)
				bot.Send(msg)
				continue
			}
			b.Users[chatid].PixelArtWidth = width
			msg.Text = fmt.Sprintf("Выберите высоту арта(в пикселях, от 8 до %d)", canvasHeigth-b.Users[chatid].DrawPixelArt.Y)
			b.Users[chatid].Stage = choseeHeight
			bot.Send(msg)
		} else if b.Users[chatid].Stage == choseeHeight {
			height, err := strconv.Atoi(update.Message.Text)
			if err != nil {
				msg.Text = "Введите целое число"
				bot.Send(msg)
				continue
			}
			if height > canvasHeigth-b.Users[chatid].DrawPixelArt.X || height < 8 {
				msg.Text = fmt.Sprintf("Введите число от 8 до %d)", canvasHeigth-b.Users[chatid].DrawPixelArt.Y)
				bot.Send(msg)
				continue
			}
			b.Users[chatid].PixelArtHeight = height
			b.Users[chatid].Stage = pixelArt
			msg.Text = "Загрузите изображение формата jpeg или PNG"
			bot.Send(msg)
		} else if b.Users[chatid].Stage == pixelArt {
			if b.Users[chatid].Stage == pixelArt {
				// Handle both photo and document uploads
				if update.Message.Photo != nil && len(update.Message.Photo) > 0 {
					photo := update.Message.Photo[len(update.Message.Photo)-1]
					fileConfig := tgbotapi.FileConfig{
						FileID: photo.FileID,
					}
					file, err := bot.GetFile(fileConfig)
					if err != nil {
						logrus.Error(err)
						continue
					}
					format := "png" // Default format

					err = b.service.Converter.Download(file.Link(bot.Token), fmt.Sprintf("images/%d.%s", update.Message.Chat.ID, format))
					if err != nil {
						logrus.Error(err)
						continue
					}
					pixel := make(chan *tgbot.Pixel, channelBuffer)
					ctx := context.Background()
					ctx, cancel := context.WithCancel(ctx)
					cancel()
					bytes, err := b.service.Pixelize("images/", fmt.Sprintf("%d.%s", update.Message.Chat.ID, format), b.Users[chatid].PixelArtWidth, b.Users[chatid].PixelArtHeight, pixel, ctx, b.Users[chatid].ID)
					if err != nil {
						logrus.Error(err)
						continue
					}
					photoFilyBytes := tgbotapi.FileBytes{
						Name:  file.FileID,
						Bytes: bytes,
					}
					msg := tgbotapi.NewPhoto(update.Message.Chat.ID, photoFilyBytes)
					msg.Caption = "ПиксельАрт будет иметь такой вид"
					msg.ReplyMarkup = Review
					bot.Send(msg)
					b.Users[chatid].Stage = review
					continue
				} else if update.Message.Document != nil && update.Message.Document.FileID != "" {
					document := update.Message.Document
					fileConfig := tgbotapi.FileConfig{
						FileID: document.FileID,
					}
					file, err := bot.GetFile(fileConfig)
					if err != nil {
						logrus.Error(err)
						continue
					}
					format := strings.ToLower(strings.Split(document.FileName, ".")[1])

					// Убедитесь, что формат поддерживается
					if format != "png" && format != "pdf" {
						msg.Text = "Поддерживаемые форматы: PNG, PDF"
						bot.Send(msg)
						continue
					}

					err = b.service.Converter.Download(file.Link(bot.Token), fmt.Sprintf("images/%d.%s", update.Message.Chat.ID, format))
					if err != nil {
						logrus.Error(err)
						continue
					}
					pixel := make(chan *tgbot.Pixel, channelBuffer)
					ctx := context.Background()
					ctx, cancel := context.WithCancel(ctx)
					cancel()
					bytes, err := b.service.Pixelize("images/", fmt.Sprintf("%d.%s", update.Message.Chat.ID, format), b.Users[chatid].PixelArtWidth, b.Users[chatid].PixelArtHeight, pixel, ctx, b.Users[chatid].ID)
					if err != nil {
						logrus.Error(err)
						continue
					}
					photoFilyBytes := tgbotapi.FileBytes{
						Name:  document.FileID,
						Bytes: bytes,
					}
					msg := tgbotapi.NewPhoto(update.Message.Chat.ID, photoFilyBytes)
					msg.Caption = "ПиксельАрт будет иметь такой вид"
					msg.ReplyMarkup = Review
					bot.Send(msg)
					b.Users[chatid].Stage = review
					continue
				}
			}
		} else if b.Users[chatid].Stage == review {
			result := update.Message.Text
			if result != "Да" && result != "Нет" {
				b.Users[chatid].Stage = active
				msg.Text = "Пожалуйста, ответьте 'Да' или 'Нет'."
				bot.Send(msg)
				continue
			}
			if result == "Да" {
				err := b.service.Converter.SendPixels("images", fmt.Sprintf("%d.png", update.Message.Chat.ID), b.Users[chatid].PixelArtWidth, b.Users[chatid].PixelArtHeight, b.Users[chatid].DrawPixelArt, b.Users[chatid].ID)
				if err != nil {
					logrus.Error(err)
					msg.Text = "Ошибка при отправке пиксель-арта."
				} else {
					msg.Text = "Пиксель-арт загружен."
				}
				b.Users[chatid].Stage = active
				msg.ReplyMarkup = actionNumber
				bot.Send(msg)
				continue
			} else {
				msg.Text = "Отмена действия."
				b.Users[chatid].Stage = active
				msg.ReplyMarkup = actionNumber
				bot.Send(msg)
				continue

			}
		}
	}
}

// Функция для проверки валидности HEX цвета
