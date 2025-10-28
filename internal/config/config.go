package config

import (
	"flag"
	"os"
)

type Config struct {
	Address         string
	BaseURL         string
	FileStoragePath string // новый параметр
}

func Init() *Config {
	addressFlag := flag.String("a", "127.0.0.1:8080", "адрес запуска HTTP-сервера")
	baseURLFlag := flag.String("b", "http://localhost:8080", "базовый адрес сокращённого URL")
	fileFlag := flag.String("f", "data.json", "путь до файла для хранения URL")

	flag.Parse()

	addressEnv := os.Getenv("SERVER_ADDRESS")
	baseURLEnv := os.Getenv("BASE_URL")
	fileEnv := os.Getenv("FILE_STORAGE_PATH")

	address := *addressFlag
	baseURL := *baseURLFlag
	filePath := *fileFlag

	if addressEnv != "" {
		address = addressEnv
	}
	if baseURLEnv != "" {
		baseURL = baseURLEnv
	}
	if fileEnv != "" {
		filePath = fileEnv
	}

	return &Config{
		Address:         address,
		BaseURL:         baseURL,
		FileStoragePath: filePath,
	}
}
