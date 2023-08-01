package leds

import (
	"github.com/PerformLine/go-stockutil/colorutil"
)

func GenerateRainbow(ledArrayChan chan [][3]byte, offset float64) {
	ledArray := make([][3]byte, 97)
	for i := 0; i < 97; i++ {
		hue := float64(i) / 97 * 360
		r, g, b := colorutil.HsvToRgb(hue+offset, 1, 1)
		rgb := [3]byte{r, g, b}
		for j := 0; j < 3; j++ {
			ledArray[i][j] = rgb[j]
		}
	}
	ledArrayChan <- ledArray
}
