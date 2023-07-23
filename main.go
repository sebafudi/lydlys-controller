package main

import (
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/PerformLine/go-stockutil/colorutil"
)

func generate_rainbow(led_array_chan chan [97][3]byte, offset float64) {
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

func start_connection(ip string, port string) net.Conn {
	conn, err := net.Dial("udp", ip+":"+port)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return conn
}

func send_udp_packet(conn net.Conn, led_array [97][3]byte) {
	var bytes_to_send []byte
	for i := 0; i < 97; i++ {
		for j := 0; j < 3; j++ {
			bytes_to_send = append(bytes_to_send, led_array[i][j])
		}
	}
	_, err := conn.Write(bytes_to_send)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func main() {
	ip := *flag.String("ip", "10.45.5.32", "IP address to send UDP packets to")
	port := *flag.String("port", "4210", "Port to send UDP packets to")
	flag.Parse()
	connection := start_connection(ip, port)
	// var led_array = make([97][3]byte{})
	led_array := make(chan [97][3]byte)
	const fps = 60
	offset := 0.0
	var frame_duration time.Duration = time.Second / time.Duration(fps)
	for {
		start := time.Now()
		go generate_rainbow(led_array, offset)
		send_udp_packet(connection, <-led_array)
		offset += 1
		for time.Since(start) < frame_duration-time.Duration(time.Since(start).Milliseconds()) {
		}
		fmt.Println(time.Since(start))
	}
}
