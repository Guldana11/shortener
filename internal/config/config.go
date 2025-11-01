package config

import (
	"flag"
	"os"
)

type Config struct {
	Address         string
	BaseURL         string
	FileStoragePath string
}

func Init() *Config {
	addressFlag := flag.String("a", "127.0.0.1:8080", "адрес запуска HTTP-сервера")
	baseURLFlag := flag.String("b", "http://localhost:8080", "базовый адрес сокращённого URL")
	fileFlag := flag.String("f", "data.json", "путь до файла для хранения URL")

	flag.Parse()

	address := *addressFlag
	baseURL := *baseURLFlag
	filePath := *fileFlag

	if v, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		address = v
	}
	if v, ok := os.LookupEnv("BASE_URL"); ok {
		baseURL = v
	}
	if v, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		filePath = v
	}

	return &Config{
		Address:         address,
		BaseURL:         baseURL,
		FileStoragePath: filePath,
	}
}
