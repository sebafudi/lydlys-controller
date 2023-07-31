package connection

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/sebafudi/lydlys-controller/internal/keys"
)

func StartConnection(ip string, port string) net.Conn {
	conn, err := net.Dial("udp", ip+":"+port)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return conn
}

func RegisterDevice(serverUrl string, pub_base64 string, serial string) {
	address := serverUrl + "/registerDevice"
	data := fmt.Sprintf(`{"pub_key": "%s", "serial": "%s"}`, pub_base64, serial)
	_, err := http.Post(address, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func GetUserToken(serverUrl string, serial string, priv []byte) string {
	address := serverUrl + "/getUserToken?email=" + os.Getenv("TEST_EMAIL") + "&serial=" + serial
	resp, err := http.Get(address)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	defer resp.Body.Close()

	var bodyString string

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return ""
		}
		bodyString = string(bodyBytes)
	}

	bodyBytes, err := base64.StdEncoding.DecodeString(bodyString)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	decrypted_data := keys.Decode_using_private_key(priv, bodyBytes)
	if decrypted_data == nil {
		fmt.Println("Error decrypting data")
		return ""
	}
	return string(decrypted_data)
}
