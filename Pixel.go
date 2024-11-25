package tgbot

type Pixel struct {
	X     int    `json:"x"`
	Y     int    `json:"y"`
	ID    int    `json:"id"`
	Color string `json:"color"`
}

type PixelClick struct {
	Pixel
	Lastclick int `json:"lastclick"`
}
type Point struct {
	X int
	Y int
}
