package leds

import (
	"fmt"
	"net"
	"time"

	"github.com/PerformLine/go-stockutil/colorutil"
	"github.com/sebafudi/lydlys-controller/internal/connection"
)

func Generate_rainbow(led_array_chan chan [97][3]byte, offset float64) {
	var led_array [97][3]byte
	for i := 0; i < 97; i++ {
		hue := float64(i) / 97 * 360
		r, g, b := colorutil.HsvToRgb(hue+offset, 1, 1)
		rgb := [3]byte{r, g, b}
		for j := 0; j < 3; j++ {
			led_array[i][j] = rgb[j]
		}

	}
	led_array_chan <- led_array
}

func sweep(offset int) [97][3]byte {
	var led_array [97][3]byte
	for i := 0; i <= 97; i++ {
		led_array[offset] = [3]byte{255, 255, 255}

	}
	return led_array
}

func BootAnimation(connectionc net.Conn, bootDone chan bool) {
	const fps = 60
	var frame_duration time.Duration = time.Second / time.Duration(fps)
	frames := make(chan [97][3]byte)
	go func() {
		for i := 0; i < 97; i++ {
			led := sweep(i)
			frames <- led
		}
		close(frames)
	}()
	go func() {
		for {
			start := time.Now()
			led_array, more := <-frames
			select {
			case <-bootDone:
				return
			default:
			}
			connection.SendUdpPacket(connectionc, led_array)
			if !more {
				return
			}
			for time.Since(start) < frame_duration-time.Duration(time.Since(start).Milliseconds()) {
			}
			fmt.Println(time.Since(start))
		}
	}()
}
