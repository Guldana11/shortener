package config

import (
	"flag"
	"os"
	"reflect"
	"testing"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		env        map[string]string
		wantConfig *Config
	}{
		{
			name: "Значения по умолчанию",
			args: []string{"cmd"},
			env:  nil,
			wantConfig: &Config{
				Address:         "127.0.0.1:8080",
				BaseURL:         "http://localhost:8080",
				FileStoragePath: "data.json",
			},
		},
		{
			name: "Переопределение через флаги",
			args: []string{"cmd", "-a", "127.0.0.1:9999", "-b", "http://example.com", "-f", "urls.json"},
			env:  nil,
			wantConfig: &Config{
				Address:         "127.0.0.1:9999",
				BaseURL:         "http://example.com",
				FileStoragePath: "urls.json",
			},
		},
		{
			name: "Переопределение через переменные окружения",
			args: []string{"cmd", "-a", "127.0.0.1:9999", "-b", "http://example.com", "-f", "urls.json"},
			env: map[string]string{
				"SERVER_ADDRESS":    "0.0.0.0:3000",
				"BASE_URL":          "http://myshort.io",
				"FILE_STORAGE_PATH": "env.json",
			},
			wantConfig: &Config{
				Address:         "0.0.0.0:3000",
				BaseURL:         "http://myshort.io",
				FileStoragePath: "env.json",
			},
		},
		{
			name: "Флаг используется если ENV пустой",
			args: []string{"cmd", "-f", "flag.json"},
			env:  map[string]string{},
			wantConfig: &Config{
				Address:         "127.0.0.1:8080",
				BaseURL:         "http://localhost:8080",
				FileStoragePath: "flag.json",
			},
		},
		{
			name: "Используется значение по умолчанию, если нет ENV и флага",
			args: []string{"cmd"},
			env:  map[string]string{},
			wantConfig: &Config{
				Address:         "127.0.0.1:8080",
				BaseURL:         "http://localhost:8080",
				FileStoragePath: "data.json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldServerAddr := os.Getenv("SERVER_ADDRESS")
			oldBaseURL := os.Getenv("BASE_URL")
			oldFilePath := os.Getenv("FILE_STORAGE_PATH")
			defer func() {
				os.Setenv("SERVER_ADDRESS", oldServerAddr)
				os.Setenv("BASE_URL", oldBaseURL)
				os.Setenv("FILE_STORAGE_PATH", oldFilePath)
			}()

			os.Unsetenv("SERVER_ADDRESS")
			os.Unsetenv("BASE_URL")
			os.Unsetenv("FILE_STORAGE_PATH")
			for k, v := range tt.env {
				os.Setenv(k, v)
			}

			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = tt.args

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			got, err := Init()
			if err != nil {
				t.Fatalf("Init() returned error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.wantConfig) {
				t.Errorf("Init() = %+v, want %+v", got, tt.wantConfig)
			}
		})
	}
}
