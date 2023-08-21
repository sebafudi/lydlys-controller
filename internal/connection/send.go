package connection

import (
	"fmt"
	"net"
)

func SendUdpPacket(conn net.Conn, ledArray [][3]byte) {
	var bytesToSend []byte
	bytesToSend = append(bytesToSend, 0x02, 0x01)
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
