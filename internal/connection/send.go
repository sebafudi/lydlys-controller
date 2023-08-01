package connection

import (
	"fmt"
	"net"
)

func SendUdpPacket(conn net.Conn, ledArray [97][3]byte) {
	var bytesToSend []byte
	for i := 0; i < 97; i++ {
		for j := 0; j < 3; j++ {
			bytesToSend = append(bytesToSend, ledArray[i][j])
		}
	}
	_, err := conn.Write(bytesToSend)
	if err != nil {
		fmt.Println(err)
		return
	}
}
