package main

import (
	"fmt"
	"image"
	"time"

	"github.com/kbinani/screenshot"
	"github.com/sebafudi/lydlys-controller/internal/config"
	"github.com/sebafudi/lydlys-controller/internal/connection"
)

func main() {
	err := config.ParseEnvs()
	if err != nil {
		fmt.Println(err)
		return
	}
	flags := config.GetFlags()
	connectionc := connection.StartConnection(*flags.Ip, *flags.Port)
	screenshot.NumActiveDisplays()

	numberOfLeds := 97
	byteColor := make([][3]byte, numberOfLeds)
	width := screenshot.GetDisplayBounds(2).Dx()
	height := screenshot.GetDisplayBounds(2).Dy()
	lastTime := time.Now()
	totalFrames := 0
	for {
		totalFrames++
		if time.Since(lastTime) > time.Second {
			fmt.Println("FPS:", totalFrames)
			totalFrames = 0
			lastTime = time.Now()
		}

		var img *image.RGBA
		bounds := image.Rect(0, height/2, width, height/2+1)
		img, err = screenshot.CaptureRect(bounds)
		if err != nil {
			panic(err)
		}

		for i := 0; i < numberOfLeds; i++ {
			color := img.At(width/numberOfLeds*i, 0)
			r, g, b, _ := color.RGBA()
			byteColor[i][0] = byte(r)
			byteColor[i][1] = byte(g)
			byteColor[i][2] = byte(b)
		}

		connection.SendUdpPacket(connectionc, byteColor)

	}

}
