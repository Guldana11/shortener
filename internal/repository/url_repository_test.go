package repository

import (
	"errors"
	"sync"
	"testing"

	"github.com/Guldana11/shortener/internal/config"
)

func TestNewURLRepository(t *testing.T) {
	cfg := config.Init()

	tests := []struct {
		name string
	}{
		{
			name: "Создание нового репозитория",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewURLRepository(cfg.FileStoragePath)
			if repo == nil {
				t.Fatal("NewURLRepository вернул nil")
			}
			if repo.store == nil {
				t.Error("store должен быть инициализирован")
			}
			if len(repo.store) != 0 {
				t.Errorf("store должен быть пустым, но длина %d", len(repo.store))
			}
		})
	}
}

func TestURLRepository_Create(t *testing.T) {
	type fields struct {
		store map[string]string
	}
	type args struct {
		originalURL string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "Создание нового URL",
			fields: fields{
				store: make(map[string]string),
			},
			args: args{
				originalURL: "https://example.com",
			},
			wantErr: nil,
		},
		{
			name: "URL уже существует",
			fields: fields{
				store: map[string]string{
					"abcd1234": "https://example.com",
				},
			},
			args: args{
				originalURL: "https://example.com",
			},
			wantErr: ErrURLExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &URLRepository{
				store: tt.fields.store,
				mu:    sync.RWMutex{},
			}

			id, err := r.Create(tt.args.originalURL)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ожидали ошибку %v, получили %v", tt.wantErr, err)
			}

			if tt.wantErr != nil {
				return
			}

			if id == "" {
				t.Error("Create вернул пустой ID")
			}

			got, ok := r.Get(id)
			if !ok {
				t.Error("URL не найден после Create")
			}
			if got != tt.args.originalURL {
				t.Errorf("ожидали %v, получили %v", tt.args.originalURL, got)
			}
		})
	}
}

func TestURLRepository_CreateWithID(t *testing.T) {
	type fields struct {
		store map[string]string
	}
	type args struct {
		id          string
		originalURL string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Сохранение URL с известным ID",
			fields: fields{
				store: make(map[string]string),
			},
			args: args{
				id:          "testid",
				originalURL: "https://example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &URLRepository{
				store: tt.fields.store,
				mu:    sync.RWMutex{},
			}
			r.CreateWithID(tt.args.id, tt.args.originalURL)

			got, ok := r.Get(tt.args.id)
			if !ok {
				t.Error("URL не найден после CreateWithID")
			}
			if got != tt.args.originalURL {
				t.Errorf("ожидали %v, получили %v", tt.args.originalURL, got)
			}
		})
	}
}

func TestURLRepository_Get(t *testing.T) {
	type fields struct {
		store map[string]string
	}
	type args struct {
		id string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		want1  bool
	}{
		{
			name: "Существующий ID",
			fields: fields{
				store: map[string]string{
					"id123": "https://example.com",
				},
			},
			args: args{
				id: "id123",
			},
			want:  "https://example.com",
			want1: true,
		},
		{
			name: "Несуществующий ID",
			fields: fields{
				store: map[string]string{},
			},
			args: args{
				id: "unknown",
			},
			want:  "",
			want1: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &URLRepository{
				store: tt.fields.store,
				mu:    sync.RWMutex{},
			}
			got, got1 := r.Get(tt.args.id)
			if got != tt.want {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Get() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
