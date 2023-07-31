package connection

import (
	"fmt"
	"os"

	"github.com/sebafudi/lydlys-controller/internal/keys"
)

func ConnectToBackend(userToken chan string) {

	priv, pub := keys.Prepare_keys()
	if priv == nil || pub == nil {
		fmt.Println("Error preparing keys")
		return
	}
	_, pub_base64 := keys.Convert_to_base64(priv, pub)

	RegisterDevice(os.Getenv("SERVER_URL"), pub_base64, os.Getenv("SERIAL"))

	go func() {
		userToken <- GetUserToken(os.Getenv("SERVER_URL"), os.Getenv("SERIAL"), priv)
	}()
}
