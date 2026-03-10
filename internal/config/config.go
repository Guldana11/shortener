package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Address         string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
	AuditFile       string
	AuditURL        string
	EnableHTTPS     bool
	TrustedSubnet   string
}

type FileConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     *bool  `json:"enable_https"`
	AuditFile       string `json:"audit_file"`
	AuditURL        string `json:"audit_url"`
	TrustedSubnet   string `json:"trusted_subnet"`
}

func Init() (*Config, error) {
	// ВАЖНО: флаги с дефолтом "" (чтобы не перетирать env/json, если флаг не задан)
	addressFlag := flag.String("a", "", "адрес запуска HTTP-сервера")
	baseURLFlag := flag.String("b", "", "базовый адрес сокращённого URL")
	fileFlag := flag.String("f", "", "путь до файла для хранения URL")
	dsnFlag := flag.String("d", "", "строка подключения к базе данных")
	auditFileFlag := flag.String("audit-file", "", "путь к файлу аудита")
	auditURLFlag := flag.String("audit-url", "", "url сервера аудита")
	enableHTTPSFlag := flag.Bool("s", false, "enable HTTPS")
	trustedSubnetFlag := flag.String("t", "", "CIDR trusted subnet")

	// путь к json конфигу
	configPathShort := flag.String("c", "", "path to json config")
	configPathLong := flag.String("config", "", "path to json config")

	flag.Parse()

	set := map[string]bool{}
	flag.CommandLine.Visit(func(f *flag.Flag) {
		set[f.Name] = true
	})

	// 1) Defaults (самый низкий приоритет)
	cfg := &Config{
		Address:         "127.0.0.1:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "data.json",
		DatabaseDSN:     "",
		AuditFile:       "",
		AuditURL:        "",
		EnableHTTPS:     false,
	}

	// 2) JSON file (ниже env/flags)
	configPath := ""
	if *configPathShort != "" {
		configPath = *configPathShort
	} else if *configPathLong != "" {
		configPath = *configPathLong
	} else if v := os.Getenv("CONFIG"); v != "" {
		configPath = v
	}

	if configPath != "" {
		fc, err := loadFileConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
		applyFileConfig(cfg, fc)
	}

	// 3) ENV (выше файла, ниже флагов)
	applyEnv(cfg)

	// 4) FLAGS (самый высокий приоритет, применяем только если флаг реально задан)
	if set["a"] {
		cfg.Address = *addressFlag
	}
	if set["b"] {
		cfg.BaseURL = *baseURLFlag
	}
	if set["f"] {
		cfg.FileStoragePath = *fileFlag
	}
	if set["d"] {
		cfg.DatabaseDSN = *dsnFlag
	}
	if set["audit-file"] {
		cfg.AuditFile = *auditFileFlag
	}
	if set["audit-url"] {
		cfg.AuditURL = *auditURLFlag
	}
	if set["s"] {
		cfg.EnableHTTPS = *enableHTTPSFlag
	}
	if set["t"] {
		cfg.TrustedSubnet = *trustedSubnetFlag
	}

	return cfg, nil
}

func loadFileConfig(path string) (FileConfig, error) {
	var fc FileConfig
	b, err := os.ReadFile(path)
	if err != nil {
		return fc, fmt.Errorf("read config file: %w", err)
	}
	if err := json.Unmarshal(b, &fc); err != nil {
		return fc, fmt.Errorf("parse config json: %w", err)
	}
	return fc, nil
}

func applyFileConfig(dst *Config, fc FileConfig) {
	if fc.ServerAddress != "" {
		dst.Address = fc.ServerAddress
	}
	if fc.BaseURL != "" {
		dst.BaseURL = fc.BaseURL
	}
	if fc.FileStoragePath != "" {
		dst.FileStoragePath = fc.FileStoragePath
	}
	if fc.DatabaseDSN != "" {
		dst.DatabaseDSN = fc.DatabaseDSN
	}
	if fc.AuditFile != "" {
		dst.AuditFile = fc.AuditFile
	}
	if fc.AuditURL != "" {
		dst.AuditURL = fc.AuditURL
	}
	if fc.EnableHTTPS != nil {
		dst.EnableHTTPS = *fc.EnableHTTPS
	}
	if fc.TrustedSubnet != "" {
		dst.TrustedSubnet = fc.TrustedSubnet
	}
}

func applyEnv(dst *Config) {
	if v := os.Getenv("SERVER_ADDRESS"); v != "" {
		dst.Address = v
	}
	if v := os.Getenv("BASE_URL"); v != "" {
		dst.BaseURL = v
	}
	if v := os.Getenv("FILE_STORAGE_PATH"); v != "" {
		dst.FileStoragePath = v
	}
	if v := os.Getenv("DATABASE_DSN"); v != "" {
		dst.DatabaseDSN = v
	}
	if v := os.Getenv("AUDIT_FILE"); v != "" {
		dst.AuditFile = v
	}
	if v := os.Getenv("AUDIT_URL"); v != "" {
		dst.AuditURL = v
	}
	if v := os.Getenv("ENABLE_HTTPS"); v != "" {
		// поддержка: true/false/1/0
		if b, err := strconv.ParseBool(v); err == nil {
			dst.EnableHTTPS = b
		}
	}
	if v := os.Getenv("TRUSTED_SUBNET"); v != "" {
		dst.TrustedSubnet = v
	}
}
