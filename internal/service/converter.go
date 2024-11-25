package service

import (
	"context"
	"fmt"
	"golang.org/x/image/draw"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"sync"
	"tgbot"

	"os"
)

type ConverterService struct {
	pixel *PixelService
}

func NewConverterService(pixel *PixelService) *ConverterService {
	return &ConverterService{pixel: pixel}
}

var palette = []color.RGBA{
	// Красный (#FF0000)
	{255, 0, 0, 255},     // Основной красный
	{255, 80, 80, 255},   // Светло-красный 2
	{255, 100, 100, 255}, // Светло-красный
	{220, 20, 20, 255},   // Темно-красный 2
	{200, 0, 0, 255},     // Темно-красный
	{150, 0, 0, 255},     // Очень темный красный

	// Зеленый (#00FF00)
	{0, 255, 0, 255},     // Основной зеленый
	{80, 255, 80, 255},   // Светло-зеленый 2
	{100, 255, 100, 255}, // Светло-зеленый
	{20, 200, 20, 255},   // Темно-зеленый 2
	{0, 200, 0, 255},     // Темно-зеленый
	{0, 150, 0, 255},     // Очень темный зеленый

	// Синий (#0000FF)
	{0, 0, 255, 255},     // Основной синий
	{80, 80, 255, 255},   // Светло-синий 2
	{100, 100, 255, 255}, // Светло-синий
	{20, 20, 200, 255},   // Темно-синий 2
	{0, 0, 200, 255},     // Темно-синий
	{0, 0, 150, 255},     // Очень темный синий

	// Желтый (#FFFF00)
	{255, 255, 0, 255},   // Основной желтый
	{255, 255, 80, 255},  // Светло-желтый 2
	{255, 255, 100, 255}, // Светло-желтый
	{200, 200, 20, 255},  // Темно-желтый 2
	{200, 200, 0, 255},   // Темно-желтый
	{150, 150, 0, 255},   // Очень темный желтый

	// Магента (#FF00FF)
	{255, 0, 255, 255},   // Основная магента
	{255, 80, 255, 255},  // Светло-магента 2
	{255, 100, 255, 255}, // Светло-магента
	{200, 0, 200, 255},   // Темно-магента 2
	{150, 0, 150, 255},   // Очень темная магента

	// Циан (#00FFFF)
	{0, 255, 255, 255},   // Основной циан
	{80, 255, 255, 255},  // Светло-циан 2
	{100, 255, 255, 255}, // Светло-циан
	{20, 200, 200, 255},  // Темно-циан 2
	{0, 150, 150, 255},   // Очень темный циан

	// Черный (#000000)
	{0, 0, 0, 255},    // Основной черный
	{50, 50, 50, 255}, // Светло-черный (темно-серый)
	{30, 30, 30, 255}, // Средне-черный
	{20, 20, 20, 255}, // Темно-черный

	// Белый (#FFFFFF)
	{255, 255, 255, 255}, // Основной белый
	{240, 240, 240, 255}, // Светло-белый 2
	{230, 230, 230, 255}, // Светло-белый
	{200, 200, 200, 255}, // Темно-белый

}

func (c *ConverterService) Download(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}
func (c *ConverterService) Pixelize(filepath string, fileName string, widthCount, heightCount int, pixelCh chan *tgbot.Pixel, ctx context.Context, id int) ([]byte, error) {
	defer close(pixelCh)
	file, err := os.Open(filepath + "/" + fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	//coefx := img.Bounds().Dx() / widthCount
	//coefy := img.Bounds().Dy() / heightCount
	width := widthCount
	height := heightCount
	smallImg := image.NewRGBA(image.Rect(0, 0, width, height))

	draw.NearestNeighbor.Scale(smallImg, smallImg.Bounds(), img, img.Bounds(), draw.Over, nil)
	ycount := smallImg.Bounds().Dy()
	xcount := smallImg.Bounds().Dx()
	wgOUT := sync.WaitGroup{}
	wgOUT.Add(ycount)

	for y := 0; y < ycount; y++ {

		go func() {
			wgInner := sync.WaitGroup{}
			wgInner.Add(xcount)
			defer wgOUT.Done()
			defer wgInner.Wait()
			for x := 0; x < xcount; x++ {
				go func(xcount int, ycount int, x, y int) {
					defer wgInner.Done()
					oldColor := smallImg.RGBAAt(x, y)
					if oldColor.A == 0 {
						return
					}
					newColor := oldColor
					pixel := tgbot.Pixel{
						X:     x,
						Y:     y,
						ID:    id,
						Color: fmt.Sprintf("#%s%s%s", toHex(newColor.R), toHex(newColor.G), toHex(newColor.B)),
					}
					smallImg.Set(x, y, newColor)
					select {
					case pixelCh <- &pixel:
					case <-ctx.Done():

					}
				}(xcount, ycount, x, y)
			}
		}()

	}

	outputFile, err := os.Create(filepath + "/" + fileName)
	if err != nil {
		return nil, err
	}
	defer outputFile.Close()
	err = png.Encode(outputFile, smallImg)
	if err != nil {
		return nil, err
	}
	photoBytes, err := ioutil.ReadFile(filepath + "/" + fileName)
	if err != nil {
		return nil, err
	}
	wgOUT.Wait()

	return photoBytes, nil
}

func (c *ConverterService) SendPixels(filepath string, fileName string, widthCount, heightCount int, point tgbot.Point, id int) error {
	conn := c.pixel.ConnectToWebSocket("ws", "localhost:8080", "/webhook/ws")
	pixelCh := make(chan *tgbot.Pixel, 1)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		_, _ = c.Pixelize(filepath, fileName, widthCount, heightCount, pixelCh, ctx, id)

	}()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		pixels := make([]tgbot.PixelClick, widthCount*heightCount)
		i := 0
		for pixel := range pixelCh {
			pixels[i] = tgbot.PixelClick{
				*pixel,
				0,
			}
			i++
		}
		c.pixel.SendPixel(pixels, conn)
	}()
	wg.Wait()

	return nil
}

// Функция для нахождения ближайшего цвета из палитры
func findClosestColor(c color.RGBA) color.RGBA {
	minDistance := math.MaxFloat64
	closestColor := palette[0]

	for _, p := range palette {
		// Вычисление расстояния между цветами
		distance := colorDistance(c, p)
		if distance < minDistance {
			minDistance = distance
			closestColor = p
		}
	}

	return closestColor
}
func colorDistance(c1, c2 color.RGBA) float64 {
	rDiff := float64(c1.R) - float64(c2.R)
	gDiff := float64(c1.G) - float64(c2.G)
	bDiff := float64(c1.B) - float64(c2.B)
	return math.Sqrt(rDiff*rDiff + gDiff*gDiff + bDiff*bDiff)
}
func toHex(number uint8) string {
	Newnumber := number
	result := ""
	convertToHex := map[uint8]string{
		0:  "0",
		1:  "1",
		2:  "2",
		3:  "3",
		4:  "4",
		5:  "5",
		6:  "6",
		7:  "7",
		8:  "8",
		9:  "9",
		10: "A",
		11: "B",
		12: "C",
		13: "D",
		14: "E",
		15: "F",
	}
	for Newnumber != 0 {
		tmp := Newnumber % 16

		result = convertToHex[tmp] + result
		Newnumber /= 16
	}
	if len(result) < 1 {
		result = "0" + result
	}
	if len(result) < 2 {
		result = "0" + result
	}
	return result
}
