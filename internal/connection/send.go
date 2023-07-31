package connection

import (
	"fmt"
	"net"
)

func SendUdpPacket(conn net.Conn, led_array [97][3]byte) {
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
