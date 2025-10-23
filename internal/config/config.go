package config

import (
	"flag"
	"os"
)

type Config struct {
	Address string
	BaseURL string
}

func Init() *Config {
	addressFlag := flag.String("a", "127.0.0.1:8080", "адрес запуска HTTP-сервера")
	baseURLFlag := flag.String("b", "http://localhost:8080", "базовый адрес сокращённого URL")

	flag.Parse()

	addressEnv := os.Getenv("SERVER_ADDRESS")
	baseURLEnv := os.Getenv("BASE_URL")

	address := *addressFlag
	baseURL := *baseURLFlag

	if addressEnv != "" {
		address = addressEnv
	}
	if baseURLEnv != "" {
		baseURL = baseURLEnv
	}

	return &Config{
		Address: address,
		BaseURL: baseURL,
	}
}
