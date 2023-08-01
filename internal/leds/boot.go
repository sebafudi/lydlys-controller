package leds

import (
	"fmt"
	"net"
	"time"

	"github.com/sebafudi/lydlys-controller/internal/connection"
)

func BootAnimation(connectionc net.Conn, bootDone chan bool) {
	const fps = 60
	var frameDuration time.Duration = time.Second / time.Duration(fps)
	frames := make(chan [][3]byte, 97)
	go func() {
		for i := 0; i < 97; i++ {
			led := GenerateSweep(i)
			frames <- led
		}
		close(frames)
	}()
	go func() {
		for {
			start := time.Now()
			ledArray, more := <-frames
			select {
			case <-bootDone:
				return
			default:
			}
			connection.SendUdpPacket(connectionc, ledArray)
			if !more {
				return
			}
			for time.Since(start) < frameDuration-time.Duration(time.Since(start).Milliseconds()) {
			}
			fmt.Println(time.Since(start))
		}
	}()
}
