package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"os"
)

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

func Convert_to_base64(priv, pub []byte) (string, string) {
	priv_base64 := base64.StdEncoding.EncodeToString(priv)
	pub_base64 := base64.StdEncoding.EncodeToString(pub)
	return priv_base64, pub_base64
}

func create_new_keys() ([]byte, []byte) {
	priv, pub := create_keys()
	priv_base64, pub_base64 := Convert_to_base64(priv, pub)
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

func Prepare_keys() ([]byte, []byte) {
	var priv, pub []byte = make([]byte, 64), make([]byte, 32)

	if _, err := os.Stat("priv.key"); os.IsNotExist(err) {
		priv, pub = create_new_keys()
	} else {
		priv, pub = read_keys()
	}
	return priv, pub
}

func Decode_using_private_key(priv []byte, data []byte) []byte {
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

func Prepare_keys_base64() (string, string) {
	priv, pub := Prepare_keys()
	priv_base64, pub_base64 := Convert_to_base64(priv, pub)
	return priv_base64, pub_base64
}
