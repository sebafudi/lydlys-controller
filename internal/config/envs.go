package config

import (
	"os"

	"github.com/joho/godotenv"
)

func ParseEnvs() error {
	goEnv := os.Getenv("GO_ENV")
	if goEnv == "" || goEnv == "development" {
		err := godotenv.Load()
		return err
	}
	return nil
}
