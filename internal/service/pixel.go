package service

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"log"
	"net/url"
	"sync"
	"tgbot"
)

type PixelService struct {
}

func NewPixelService() *PixelService {
	return &PixelService{}
}

func (p *PixelService) DrawRectangel(x0, y0, x1, y1 int, color string, id int) {
	conn := p.ConnectToWebSocket("ws", "localhost:8080", "/webhook/ws")
	wgOUT := sync.WaitGroup{}
	defer wgOUT.Wait()
	go func() {
		wgOUT.Add(1)
		for x := x0; x <= x1; x++ {
			i := 0
			pixels := make([]tgbot.PixelClick, (y1 - y0 + 1))
			for y := y0; y <= y1; y++ {

				pixel := tgbot.Pixel{
					ID:    id,
					X:     x,
					Y:     y,
					Color: color,
				}
				pixels[i] = tgbot.PixelClick{pixel, 0}
				i++
			}
			p.SendPixel(pixels, conn)
		}
		wgOUT.Done()
	}()
}

func (p *PixelService) SendPixel(pixels []tgbot.PixelClick, conn *websocket.Conn) {

	JSONpixel, err := json.Marshal(pixels)
	if err != nil {
		logrus.Error(err)
	}
	err = conn.WriteMessage(websocket.TextMessage, JSONpixel)
	if err != nil {
		logrus.Error(err)
	}

}
func (p *PixelService) ConnectToWebSocket(Scheme, host, path string) *websocket.Conn {
	u := url.URL{Scheme: Scheme, Host: host, Path: path}
	logrus.Printf("connecting to %s", u.String())
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	return conn
}
