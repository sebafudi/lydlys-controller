package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sebafudi/lydlys-controller/internal/leds"
)

func main() {
	ledCount := flag.Int("ledCount", 97, "Number of leds in the strip")
	seconds := flag.Int("seconds", 3, "Number of seconds to run the animation")
	fps := flag.Int("fps", 60, "Number of frames per second")
	out := flag.String("out", "./tmp/rainbow.lys", "Output file name")
	flag.Parse()

	// ledArrayChan := make(chan [][3]byte, *ledCount)
	ledBuffer := make([][][3]byte, *seconds**fps)

	for i := range ledBuffer {
		ledBuffer[i] = make([][3]byte, *ledCount)
	}

	toWrite := make([]byte, 0)
	for i := 0; i < *fps**seconds; i++ {
		// go leds.GenerateRainbow(ledArrayChan, float64(i))
		ledBuffer := leds.GenerateSweep(i)
		// ledBuffer := <-ledArrayChan
		for j := 0; j < *ledCount; j++ {
			for k := 0; k < 3; k++ {
				toWrite = append(toWrite, ledBuffer[j][k])
			}
		}
	}

	err := os.WriteFile(*out, toWrite, 0644)
	if err != nil {
		fmt.Println(err)
	}
}
