// Package config предоставляет функционал для чтения конфигурации приложения
// из флагов командной строки и переменных окружения.
package config

import (
	"flag"
	"os"
)

// Config хранит конфигурацию сервера и хранилища URL.
type Config struct {
	// Address — адрес запуска HTTP-сервера, например "127.0.0.1:8080".
	Address string
	// BaseURL — базовый адрес сокращённого URL, например "http://localhost:8080".
	BaseURL string
	// FileStoragePath — путь до файла для хранения URL (если используется файловое хранилище).
	FileStoragePath string
	// DatabaseDSN — строка подключения к базе данных.
	DatabaseDSN string
}

// Init инициализирует конфигурацию приложения.
// Сначала читает значения из флагов командной строки, затем из переменных окружения.
// Возвращает указатель на Config.
func Init() *Config {
	addressFlag := flag.String("a", "127.0.0.1:8080", "адрес запуска HTTP-сервера")
	baseURLFlag := flag.String("b", "http://localhost:8080", "базовый адрес сокращённого URL")
	fileFlag := flag.String("f", "data.json", "путь до файла для хранения URL")
	dsnFlag := flag.String("d", "", "строка подключения к базе данных")
	auditFileFlag := flag.String("audit-file", "", "путь к файлу аудита")
	auditURLFlag := flag.String("audit-url", "", "url сервера аудита")

	flag.Parse()

	address := *addressFlag
	baseURL := *baseURLFlag
	filePath := *fileFlag
	dsn := *dsnFlag
	auditFile := *auditFileFlag
	auditURL := *auditURLFlag

	if v, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		address = v
	}
	if v, ok := os.LookupEnv("BASE_URL"); ok {
		baseURL = v
	}
	if v, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		filePath = v
	}
	if v, ok := os.LookupEnv("DATABASE_DSN"); ok {
		dsn = v
	}
	if v, ok := os.LookupEnv("AUDIT_FILE"); ok {
		auditFile = v
	}
	if v, ok := os.LookupEnv("AUDIT_URL"); ok {
		auditURL = v
	}

	return &Config{
		Address:         address,
		BaseURL:         baseURL,
		FileStoragePath: filePath,
		DatabaseDSN:     dsn,
		AuditFile:       auditFile,
		AuditURL:        auditURL,
	}
}
