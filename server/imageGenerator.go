package main

import (
	"image"
	"image/png"
	"os"
	"strconv"
)

func main() {
	for i := 1; i <= 20; i++ {
		createImage(i*1000, i*1000, "image"+strconv.Itoa(i))
	}

}

func createImage(width int, height int, name string) {

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Colors are defined by Red, Green, Blue, Alpha uint8 values.
	//cyan := color.RGBA{100, 200, 200, 0xff}

	// Set color for each pixel.
	//for x := 0; x < width; x++ {
	//	for y := 0; y < height; y++ {
	//		switch {
	//		case x < width/2 && y < height/2: // upper left quadrant
	//			img.Set(x, y, cyan)
	//		case x >= width/2 && y >= height/2: // lower right quadrant
	//			img.Set(x, y, color.White)
	//		default:
	//			// Use zero value.
	//		}
	//	}
	//}

	// Encode as PNG.
	f, _ := os.Create(name + ".png")
	png.Encode(f, img)
}
