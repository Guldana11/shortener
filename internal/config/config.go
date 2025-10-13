package config

import (
	"flag"
)

type Config struct {
	Address string
	BaseURL string
}

func Init() *Config {
	address := flag.String("a", "localhost:8080", "адрес запуска HTTP-сервера")
	baseURL := flag.String("b", "http://localhost:8080", "базовый адрес сокращённого URL")

	flag.Parse()

	return &Config{
		Address: *address,
		BaseURL: *baseURL,
	}
}
