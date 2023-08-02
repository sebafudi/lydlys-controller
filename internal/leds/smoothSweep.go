package leds

import (
	"math"
)

func white(brightness byte) [3]byte {
	return [3]byte{brightness, brightness, brightness}
}

func GenerateSmoothSweep(numberOfLeds int, numberOfFrames int) [][][3]byte {
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
			returnArray[i][j] = white(brightness)
		}
	}
	return returnArray
}
