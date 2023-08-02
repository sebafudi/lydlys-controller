package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/sebafudi/lydlys-controller/internal/leds"
)

func main() {
	ledCount := flag.Int("ledCount", 97, "Number of leds in the strip")
	seconds := flag.Int("seconds", 3, "Number of seconds to run the animation")
	fps := flag.Int("fps", 60, "Number of frames per second")
	out := flag.String("out", "./tmp/rainbow.lys", "Output file name")
	color := flag.String("color", "white", "Color to use for the animation")
	flag.Parse()

	ledBuffer := make([][][3]byte, *seconds**fps)

	for i := range ledBuffer {
		ledBuffer[i] = make([][3]byte, *ledCount)
	}

	toWrite := make([]byte, 0)

	// GENERATE LED BUFFER
	startTime := time.Now()
	var totalTime time.Duration
	ledBuffer = leds.GenerateSmoothSweep(*ledCount, *seconds**fps, *color)
	totalTime = time.Since(startTime)
	fmt.Printf("Time to generate: %v\t %v per frame\n", totalTime, totalTime/time.Duration(*seconds**fps))

	// APPEND TO BYTE SLICE
	startTime = time.Now()
	for i := 0; i < *seconds**fps; i++ {
		for j := 0; j < *ledCount; j++ {
			for k := 0; k < 3; k++ {
				toWrite = append(toWrite, ledBuffer[i][j][k])
			}
		}
	}
	totalTime += time.Since(startTime)
	fmt.Printf("Time to append: %v\t total: %v\n", time.Since(startTime), totalTime)

	// WRITE TO FILE
	startTime = time.Now()
	err := os.WriteFile(*out, toWrite, 0644)
	if err != nil {
		fmt.Println(err)
	}
	totalTime += time.Since(startTime)
	fmt.Printf("Time to write: %v\t total: %v\n", time.Since(startTime), totalTime)
}
