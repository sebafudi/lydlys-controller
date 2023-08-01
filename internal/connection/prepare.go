package connection

import (
	"fmt"
	"os"

	"github.com/sebafudi/lydlys-controller/internal/keys"
)

func ConnectToBackend(userToken chan string) {

	priv, pub := keys.PrepareKeys()
	if priv == nil || pub == nil {
		fmt.Println("Error preparing keys")
		return
	}
	_, pubBase64 := keys.ConvertToBase64(priv, pub)

	RegisterDevice(os.Getenv("SERVER_URL"), pubBase64, os.Getenv("SERIAL"))

	go func() {
		userToken <- GetUserToken(os.Getenv("SERVER_URL"), os.Getenv("SERIAL"), priv)
	}()
}
