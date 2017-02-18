package main

import (
	"image/png"
	"os"

	"github.com/jbowens/postcards2diane"
)

func main() {
	p := postcards2diane.New("6x11", []string{"IMPEACH", "DONALD", "TRUMP"})

	img, err := p.Render()
	if err != nil {
		panic(err)
	}
	f, err := os.Create("/tmp/diane.png")
	if err != nil {
		panic(err)
	}
	err = png.Encode(f, img)
	if err != nil {
		panic(err)
	}
}
