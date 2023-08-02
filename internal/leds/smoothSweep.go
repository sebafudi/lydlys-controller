package leds

import (
	"math"
)

func colorSwitch(brightness byte, color string) [3]byte {
	switch color {
	case "red":
		return [3]byte{brightness, 0, 0}
	case "green":
		return [3]byte{0, brightness, 0}
	case "blue":
		return [3]byte{0, 0, brightness}
	case "yellow":
		return [3]byte{brightness, brightness, 0}
	case "cyan":
		return [3]byte{0, brightness, brightness}
	case "magenta":
		return [3]byte{brightness, 0, brightness}
	case "white":
		return [3]byte{brightness, brightness, brightness}
	default:
		return [3]byte{brightness, brightness, brightness}
	}
}

func GenerateSmoothSweep(numberOfLeds int, numberOfFrames int, color string) [][][3]byte {
	returnArray := make([][][3]byte, numberOfFrames)
	root1 := (math.Sqrt(-4 * -255)) / (2 * 1)
	zeroOffset := int(math.Round(math.Abs(root1)))
	for i := 0; i < numberOfFrames; i++ {
		returnArray[i] = make([][3]byte, numberOfLeds)
		for j := 0; j < numberOfLeds; j++ {
			a := ((j % numberOfLeds) - (i % (numberOfLeds + zeroOffset*2)) + zeroOffset)
			const c = 255
			value := -a*a + c
			if value < 0 {
				value = 0
			}
			brightness := byte(value)
			returnArray[i][j] = colorSwitch(brightness, color)
		}
	}
	return returnArray
}
