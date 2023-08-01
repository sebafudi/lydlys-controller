package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"os"
)

func createKeys() ([]byte, []byte) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	pub := priv.Public()
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	return privBytes, pubBytes
}

func ConvertToBase64(priv, pub []byte) (string, string) {
	privBase64 := base64.StdEncoding.EncodeToString(priv)
	pubBase64 := base64.StdEncoding.EncodeToString(pub)
	return privBase64, pubBase64
}

func createNewKeys() ([]byte, []byte) {
	priv, pub := createKeys()
	privBase64, pubBase64 := ConvertToBase64(priv, pub)
	os.WriteFile("pub.key", []byte(pubBase64), 0644)
	os.WriteFile("priv.key", []byte(privBase64), 0644)
	return priv, pub
}

func readKeys() ([]byte, []byte) {
	privBase64, err := os.ReadFile("priv.key")
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	pubBase64, err := os.ReadFile("pub.key")
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	priv, err := base64.StdEncoding.DecodeString(string(privBase64))
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	pub, err := base64.StdEncoding.DecodeString(string(pubBase64))
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	return priv, pub
}

func PrepareKeys() ([]byte, []byte) {
	var priv, pub []byte = make([]byte, 64), make([]byte, 32)

	if _, err := os.Stat("priv.key"); os.IsNotExist(err) {
		priv, pub = createNewKeys()
	} else {
		priv, pub = readKeys()
	}
	return priv, pub
}

func DecodeUsingPrivateKey(priv []byte, data []byte) []byte {
	privKey, err := x509.ParsePKCS1PrivateKey(priv)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	decryptedData, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, data)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return decryptedData
}

func PrepareKeysBase64() (string, string) {
	priv, pub := PrepareKeys()
	privBase64, pubBase64 := ConvertToBase64(priv, pub)
	return privBase64, pubBase64
}
