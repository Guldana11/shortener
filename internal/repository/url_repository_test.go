package repository

import (
	"sync"
	"testing"
)

func TestNewURLRepository(t *testing.T) {
	repo := NewURLRepository("")
	if repo == nil {
		t.Fatal("NewURLRepository вернул nil")
	}
	if repo.store == nil {
		t.Error("store должен быть инициализирован")
	}
	if len(repo.store) != 0 {
		t.Errorf("store должен быть пустым, но длина %d", len(repo.store))
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
			if err != tt.wantErr {
				t.Fatalf("ожидали ошибку %v, получили %v", tt.wantErr, err)
			}

			if tt.wantErr != nil {
				return
			}

			if id == "" {
				t.Error("Create вернул пустой ID")
			}

			got, err := r.Get(id)
			if err != nil {
				t.Errorf("URL не найден после Create: %v", err)
			}
			if got != tt.originalURL {
				t.Errorf("ожидали %v, получили %v", tt.originalURL, got)
			}
		})
	}
}

func TestURLRepository_CreateWithID(t *testing.T) {
	r := &URLRepository{
		store: map[string]map[string]string{
			globalUserID: {},
		},
		mu: sync.RWMutex{},
	}
	id := "testid"
	url := "https://example.com"
	r.CreateWithID(id, url)

	got, err := r.Get(id)
	if err != nil {
		t.Errorf("URL не найден после CreateWithID: %v", err)
	}
	if got != url {
		t.Errorf("ожидали %v, получили %v", url, got)
	}
}

func TestURLRepository_Get(t *testing.T) {
	r := &URLRepository{
		store: map[string]map[string]string{
			globalUserID: {
				"id123": "https://example.com",
			},
		},
		mu: sync.RWMutex{},
	}

	got, err := r.Get("id123")
	if err != nil {
		t.Errorf("Get() вернул ошибку для существующего ID: %v", err)
	}
	if got != "https://example.com" {
		t.Errorf("Get() = %v, want %v", got, "https://example.com")
	}

	got, err = r.Get("unknown")
	if err == nil {
		t.Errorf("ожидали ошибку для несуществующего ID, получили nil")
	}
	if got != "" {
		t.Errorf("ожидали пустую строку, получили %v", got)
	}
}

func TestURLRepository_MarkAsDeleted(t *testing.T) {
	r := &URLRepository{
		store: map[string]map[string]string{
			"user1": {
				"id1": "https://example.com",
				"id2": "https://example2.com",
			},
		},
		mu: sync.RWMutex{},
	}

	err := r.MarkAsDeleted("user1", []string{"id1"})
	if err != nil {
		t.Errorf("MarkAsDeleted вернул ошибку: %v", err)
	}

	if _, err := r.Get("id1"); err == nil {
		t.Errorf("URL id1 должен быть удалён")
	}
	if _, err := r.Get("id2"); err != nil {
		t.Errorf("URL id2 не должен быть удалён: %v", err)
	}
}
