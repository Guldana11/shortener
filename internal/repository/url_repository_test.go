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
			if len(repo.store) != 0 && len(repo.store[globalUserID]) != 0 {
				t.Errorf("store должен быть пустым, но длина %d", len(repo.store[globalUserID]))
			}
		})
	}
}

func TestURLRepository_Create(t *testing.T) {
	tests := []struct {
		name        string
		store       map[string]map[string]string
		originalURL string
		wantErr     error
	}{
		{
			name: "Создание нового URL",
			store: map[string]map[string]string{
				globalUserID: {},
			},
			originalURL: "https://example.com",
			wantErr:     nil,
		},
		{
			name: "URL уже существует",
			store: map[string]map[string]string{
				globalUserID: {
					"abcd1234": "https://example.com",
				},
			},
			originalURL: "https://example.com",
			wantErr:     ErrURLExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &URLRepository{
				store: tt.store,
				mu:    sync.RWMutex{},
			}

			id, err := r.Create(tt.originalURL)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ожидали ошибку %v, получили %v", tt.wantErr, err)
			}

			if tt.wantErr != nil {
				return
			}

			if id == "" {
				t.Error("Create вернул пустой ID")
			}

			got, ok, deleted := r.Get(id)
			if !ok {
				t.Error("URL не найден после Create")
			}
			if got != tt.originalURL {
				t.Errorf("ожидали %v, получили %v", tt.originalURL, got)
			}
			if deleted {
				t.Fatalf("URL с ID %s помечен как удалённый, хотя только что создан", id)
			}
		})
	}
}

func TestURLRepository_CreateWithID(t *testing.T) {
	tests := []struct {
		name        string
		store       map[string]map[string]string
		id          string
		originalURL string
	}{
		{
			name: "Сохранение URL с известным ID",
			store: map[string]map[string]string{
				globalUserID: {},
			},
			id:          "testid",
			originalURL: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &URLRepository{
				store: tt.store,
				mu:    sync.RWMutex{},
			}
			r.CreateWithID(tt.id, tt.originalURL)

			got, ok, deleted := r.Get(tt.id)
			if !ok {
				t.Error("URL не найден после CreateWithID")
			}
			if got != tt.originalURL {
				t.Errorf("ожидали %v, получили %v", tt.originalURL, got)
			}
			if deleted {
				t.Fatalf("URL с ID %s помечен как удалённый, хотя только что создан", tt.id)
			}
		})
	}
}

func TestURLRepository_Get(t *testing.T) {
	tests := []struct {
		name  string
		store map[string]map[string]string
		id    string
		want  string
		want1 bool
	}{
		{
			name: "Существующий ID",
			store: map[string]map[string]string{
				globalUserID: {
					"id123": "https://example.com",
				},
			},
			id:    "id123",
			want:  "https://example.com",
			want1: true,
		},
		{
			name: "Несуществующий ID",
			store: map[string]map[string]string{
				globalUserID: {},
			},
			id:    "unknown",
			want:  "",
			want1: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &URLRepository{
				store: tt.store,
				mu:    sync.RWMutex{},
			}
			got, got1, deleted := r.Get(tt.id)
			if got != tt.want {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Get() got1 = %v, want %v", got1, tt.want1)
			}
			if deleted {
				t.Fatalf("URL с ID %s помечен как удалённый, хотя только что создан", tt.id)
			}
		})
	}
}
