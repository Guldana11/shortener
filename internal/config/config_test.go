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
				Address: "127.0.0.1:8080",
				BaseURL: "http://localhost:8080",
			},
		},
		{
			name: "Переопределение через флаги",
			args: []string{"cmd", "-a", "127.0.0.1:9999", "-b", "http://example.com"},
			env:  nil,
			wantConfig: &Config{
				Address: "127.0.0.1:9999",
				BaseURL: "http://example.com",
			},
		},
		{
			name: "Переопределение через переменные окружения",
			args: []string{"cmd", "-a", "127.0.0.1:9999", "-b", "http://example.com"},
			env: map[string]string{
				"SERVER_ADDRESS": "0.0.0.0:3000",
				"BASE_URL":       "http://myshort.io",
			},
			wantConfig: &Config{
				Address: "0.0.0.0:3000",
				BaseURL: "http://myshort.io",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldServerAddr := os.Getenv("SERVER_ADDRESS")
			oldBaseURL := os.Getenv("BASE_URL")
			defer func() {
				os.Setenv("SERVER_ADDRESS", oldServerAddr)
				os.Setenv("BASE_URL", oldBaseURL)
			}()

			os.Unsetenv("SERVER_ADDRESS")
			os.Unsetenv("BASE_URL")
			for k, v := range tt.env {
				os.Setenv(k, v)
			}

			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = tt.args

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			got := Init()
			if !reflect.DeepEqual(got, tt.wantConfig) {
				t.Errorf("Init() = %+v, want %+v", got, tt.wantConfig)
			}
		})
	}
}
