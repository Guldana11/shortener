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
		wantConfig *Config
	}{
		{
			name: "Значения по умолчанию",
			args: []string{"cmd"},
			wantConfig: &Config{
				Address: "127.0.0.1:8080",
				BaseURL: "http://localhost:8080",
			},
		},
		{
			name: "Переопределение через флаги",
			args: []string{"cmd", "-a", "127.0.0.1:9999", "-b", "http://example.com"},
			wantConfig: &Config{
				Address: "127.0.0.1:9999",
				BaseURL: "http://example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
