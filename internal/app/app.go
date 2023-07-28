package app

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/PerformLine/go-stockutil/colorutil"
	"github.com/joho/godotenv"
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

func create_keys() ([]byte, []byte) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	pub := priv.Public()
	priv_bytes := x509.MarshalPKCS1PrivateKey(priv)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	pub_bytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	return priv_bytes, pub_bytes
}

func convert_to_base64(priv, pub []byte) (string, string) {
	priv_base64 := base64.StdEncoding.EncodeToString(priv)
	pub_base64 := base64.StdEncoding.EncodeToString(pub)
	return priv_base64, pub_base64
}

func create_new_keys() ([]byte, []byte) {
	priv, pub := create_keys()
	priv_base64, pub_base64 := convert_to_base64(priv, pub)
	os.WriteFile("pub.key", []byte(pub_base64), 0644)
	os.WriteFile("priv.key", []byte(priv_base64), 0644)
	return priv, pub
}

func read_keys() ([]byte, []byte) {
	priv_base64, err := os.ReadFile("priv.key")
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	pub_base64, err := os.ReadFile("pub.key")
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	priv, err := base64.StdEncoding.DecodeString(string(priv_base64))
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	pub, err := base64.StdEncoding.DecodeString(string(pub_base64))
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	return priv, pub
}

func prepare_keys() ([]byte, []byte) {
	var priv, pub []byte = make([]byte, 64), make([]byte, 32)

	if _, err := os.Stat("priv.key"); os.IsNotExist(err) {
		priv, pub = create_new_keys()
	} else {
		priv, pub = read_keys()
	}
	return priv, pub
}

func decode_using_private_key(priv []byte, data []byte) []byte {
	priv_key, err := x509.ParsePKCS1PrivateKey(priv)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	decrypted_data, err := rsa.DecryptPKCS1v15(rand.Reader, priv_key, data)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return decrypted_data
}

func RunApp() {

	goEnv := os.Getenv("GO_ENV")
	if goEnv == "" || goEnv == "development" {
		err := godotenv.Load()
		if err != nil {
			fmt.Println("Error loading .env file")
			return
		}
	}
	priv, pub := prepare_keys()
	if priv == nil || pub == nil {
		fmt.Println("Error preparing keys")
		return
	}
	_, pub_base64 := convert_to_base64(priv, pub)

	// send http request to server with pub key and serial number

	address := "http://localhost:3000/registerDevice"

	// send http request to server with pub key and serial number as json
	data := fmt.Sprintf(`{"pub_key": "%s", "serial": "%s"}`, string(pub_base64), os.Getenv("SERIAL"))
	_, err := http.Post(address, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		fmt.Println(err)
		return
	}

	address = fmt.Sprintf("http://localhost:3000/getUserToken?email=%s&serial=%s", os.Getenv("TEST_EMAIL"), os.Getenv("SERIAL"))
	resp, err := http.Get(address)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	var bodyString string

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		bodyString = string(bodyBytes)
	}

	bodyBytes, err := base64.StdEncoding.DecodeString(bodyString)
	if err != nil {
		fmt.Println(err)
		return
	}

	decrypted_data := decode_using_private_key(priv, bodyBytes)
	if decrypted_data == nil {
		fmt.Println("Error decrypting data")
		return
	}
	fmt.Println(string(decrypted_data))

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
