package app

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sebafudi/lydlys-controller/internal/config"
	"github.com/sebafudi/lydlys-controller/internal/connection"
	"github.com/sebafudi/lydlys-controller/internal/leds"
)

func RunApp() {
	config.LoadEnv()
	flag.Parse()
	ip := *flag.String("ip", os.Getenv("DEFAULT_IP"), "IP address to send UDP packets to")
	port := *flag.String("port", os.Getenv("DEFAULT_PORT"), "Port to send UDP packets to")
	connectionc := connection.StartConnection(ip, port)
	var wg sync.WaitGroup

	bootDone := make(chan bool)
	go func() {
		done := make(chan bool)
		go leds.BootAnimation(connectionc, done, bootDone)
		<-done
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
	fmt.Println("Done")

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
		fmt.Println(time.Since(start))
	}
}
