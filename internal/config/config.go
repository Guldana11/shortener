package config

import (
	"flag"
	"os"
)

type Config struct {
	Address         string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
	AuditFile       string
	AuditURL        string
}

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
