package app

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sebafudi/lydlys-controller/internal/config"
	"github.com/sebafudi/lydlys-controller/internal/connection"
	"github.com/sebafudi/lydlys-controller/internal/leds"
)

func RunApp() {
	err := config.ParseEnvs()
	if err != nil {
		fmt.Println(err)
		return
	}
	flags := config.GetFlags()
	connectionc := connection.StartConnection(flags.Ip, flags.Port)
	var wg sync.WaitGroup

	bootDone := make(chan bool)
	go func() {
		go leds.BootAnimation(connectionc, bootDone)
	}()

	wg.Add(1)
	userToken := make(chan string)
	go connection.ConnectToBackend(userToken)
	go func() {
		defer wg.Done()
		token := <-userToken
		fmt.Println(token)
	}()

	wg.Wait()
	bootDone <- true

	file, err := os.ReadFile("./tmp/rainbow.lys")
	if err != nil {
		fmt.Println(err)
		return
	}
	const fps = 60
	// offset := 0.0
	var frameDuration time.Duration = time.Second / time.Duration(fps)
	// ledArray := make(chan [][3]byte, 97)
	// for {
	// 	start := time.Now()
	// 	go leds.GenerateRainbow(ledArray, offset)
	// 	connection.SendUdpPacket(connectionc, <-ledArray)
	// 	offset += 1
	// 	for time.Since(start) < frameDuration-time.Duration(time.Since(start).Milliseconds()) {
	// 	}
	// }
	sinceStart := time.Now()
	lastFrame := 0
	avgDuration := time.Duration(10 * time.Second)
	skippedFrames := 0
	for {
		start := time.Now()

		frameNumber := int(time.Since(sinceStart).Seconds() * float64(fps))
		if frameNumber*(97*3) >= len(file) {
			fmt.Println("end of file")
			fmt.Printf("Duration: %v\n", time.Since(sinceStart))
			avgDuration = (avgDuration + time.Since(sinceStart)) / 2
			fmt.Printf("Avg Duration offset: %v\n", avgDuration-time.Duration(10*time.Second))
			fmt.Printf("Skipped frames: %v\n", skippedFrames)
			sinceStart = time.Now()
			frameNumber = 0
			lastFrame = 0
		}
		if frameNumber > lastFrame+1 {
			skippedFrames += frameNumber - lastFrame - 1
		}
		if frameNumber == lastFrame {
			continue
		}
		lastFrame = frameNumber
		read := file[frameNumber*97*3 : (frameNumber+1)*97*3]

		var ledBuffer [][3]byte
		for j := 0; j < 97; j++ {
			ledBuffer = append(ledBuffer, [3]byte{})
			for k := 0; k < 3; k++ {
				ledBuffer[j][k] = read[j*3+k]
			}
		}
		connection.SendUdpPacket(connectionc, ledBuffer)
		for time.Since(start) < (frameDuration-time.Duration(time.Since(start).Milliseconds()))/2 {
		}

	}
}
