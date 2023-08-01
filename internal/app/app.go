package app

import (
	"fmt"
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

	const fps = 60
	offset := 0.0
	var frame_duration time.Duration = time.Second / time.Duration(fps)
	led_array := make(chan [97][3]byte)
	for {
		start := time.Now()
		go leds.Generate_rainbow(led_array, offset)
		connection.SendUdpPacket(connectionc, <-led_array)
		offset += 1
		for time.Since(start) < frame_duration-time.Duration(time.Since(start).Milliseconds()) {
		}
	}
}
