package repository

import (
	"sync"
	"testing"

	"github.com/Guldana11/shortener/internal/model"
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
	repo := &URLRepository{
		store: map[string]map[string]string{},
		mu:    sync.RWMutex{},
	}

	id, err := repo.Create("https://example.com")
	if err != nil {
		t.Fatalf("Create вернул ошибку: %v", err)
	}

	if id == "" {
		t.Error("Create вернул пустой ID")
	}

	got, err := repo.Get(id)
	if err != nil {
		t.Errorf("URL не найден после Create: %v", err)
	}
	if got != "https://example.com" {
		t.Errorf("ожидали %v, получили %v", "https://example.com", got)
	}
}

func TestURLRepository_CreateForUser(t *testing.T) {
	userID := "user1"
	r := &URLRepository{
		store: map[string]map[string]string{},
		mu:    sync.RWMutex{},
	}

	id, err := r.CreateForUser(userID, "https://example.com")
	if err != nil {
		t.Fatalf("CreateForUser вернул ошибку: %v", err)
	}
	if id == "" {
		t.Error("CreateForUser вернул пустой ID")
	}

	got, err := r.GetForUser(userID, id)
	if err != nil {
		t.Errorf("URL не найден после CreateForUser: %v", err)
	}
	if got != "https://example.com" {
		t.Errorf("ожидали %v, получили %v", "https://example.com", got)
	}

	_, err = r.CreateForUser(userID, "https://example.com")
	if err != ErrURLExists {
		t.Errorf("ожидали ErrURLExists, получили %v", err)
	}
}

func TestURLRepository_BatchCreateForUser(t *testing.T) {
	userID := "user1"
	r := &URLRepository{
		store: map[string]map[string]string{},
		mu:    sync.RWMutex{},
	}

	urls := []string{"https://a.com", "https://b.com"}
	ids := r.BatchCreateForUser(userID, urls)

	if len(ids) != len(urls) {
		t.Fatalf("ожидали %d ID, получили %d", len(urls), len(ids))
	}

	for i, id := range ids {
		got, err := r.GetForUser(userID, id)
		if err != nil {
			t.Errorf("URL не найден после BatchCreateForUser: %v", err)
		}
		if got != urls[i] {
			t.Errorf("ожидали %v, получили %v", urls[i], got)
		}
	}
}

func TestURLRepository_GetAllForUser(t *testing.T) {
	userID := "user1"
	r := &URLRepository{
		store: map[string]map[string]string{
			userID: {
				"id1": "https://a.com",
				"id2": "https://b.com",
			},
		},
		mu: sync.RWMutex{},
	}

	all := r.GetAllForUser(userID)
	if len(all) != 2 {
		t.Errorf("ожидали 2 URL, получили %d", len(all))
	}
	if all["id1"] != "https://a.com" || all["id2"] != "https://b.com" {
		t.Errorf("данные URL не совпадают с ожидаемыми")
	}
}

func TestURLRepository_MarkAsDeleted(t *testing.T) {
	userID := "user1"
	r := &URLRepository{
		store: map[string]map[string]string{
			userID: {
				"id1": "https://a.com",
				"id2": "https://b.com",
			},
		},
		mu: sync.RWMutex{},
	}

	err := r.MarkAsDeleted(userID, []string{"id1"})
	if err != nil {
		t.Errorf("MarkAsDeleted вернул ошибку: %v", err)
	}

	if _, err := r.GetForUser(userID, "id1"); err == nil {
		t.Errorf("URL id1 должен быть удалён")
	}
	if _, err := r.GetForUser(userID, "id2"); err != nil {
		t.Errorf("URL id2 не должен быть удалён: %v", err)
	}
}

func TestURLRepository_Get_NonExistent(t *testing.T) {
	userID := "user1"
	r := &URLRepository{
		store: map[string]map[string]string{},
		mu:    sync.RWMutex{},
	}

	_, err := r.GetForUser(userID, "unknown")
	if err != model.ErrNotFound {
		t.Errorf("ожидали ErrNotFound для несуществующего ID, получили %v", err)
	}
}
