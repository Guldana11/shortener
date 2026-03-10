// Package repository содержит реализации хранилищ для сервиса сокращения URL.
// В данном файле реализовано in-memory хранилище с сохранением в файл.
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/Guldana11/shortener/internal/model"
)

// ErrURLExists возвращается, если URL уже существует для пользователя.
var ErrURLExists = errors.New("url already exists")

// URLRepository представляет in-memory хранилище URL с поддержкой пользователей.
//
// Данные хранятся в map[userID]map[id]originalURL.
// При задании пути файла store сохраняется на диск и загружается при старте.
type URLRepository struct {
	store map[string]map[string]string
	mu    sync.RWMutex
	file  string
}

// NewURLRepository создаёт новый экземпляр URLRepository.
// filePath — путь для сохранения данных на диск (может быть пустым).
func NewURLRepository(filePath string) *URLRepository {
	r := &URLRepository{
		store: make(map[string]map[string]string),
		file:  filePath,
	}
	r.loadFromFile()
	return r
}

// Create создаёт короткий URL без привязки к пользователю.
func (r *URLRepository) Create(originalURL string) (string, error) {
	return r.CreateForUser(" ", originalURL)
}

// CreateWithID создаёт URL с заданным ID без привязки к пользователю.
func (r *URLRepository) CreateWithID(id string, originalURL string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.store[" "] == nil {
		r.store[" "] = make(map[string]string)
	}

	r.store[" "][id] = originalURL
	r.saveToFile()
}

// CreateForUser создаёт короткий URL для конкретного пользователя.
//
// Если URL уже существует, возвращает ErrURLExists и существующий ID.
func (r *URLRepository) CreateForUser(userID, originalURL string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.store[userID] == nil {
		r.store[userID] = make(map[string]string)
	}

	for id, url := range r.store[userID] {
		if url == originalURL {
			return id, ErrURLExists
		}
	}

	id := generateID()
	r.store[userID][id] = originalURL
	r.saveToFile()
	return id, nil
}

// BatchCreateForUser создаёт несколько URL для пользователя и возвращает их ID.
func (r *URLRepository) BatchCreateForUser(userID string, urls []string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.store[userID] == nil {
		r.store[userID] = make(map[string]string)
	}

	ids := make([]string, len(urls))
	for i, u := range urls {
		var id string
		for {
			id = generateID()
			if _, exists := r.store[userID][id]; !exists {
				break
			}
		}
		r.store[userID][id] = u
		ids[i] = id
	}

	r.saveToFile()
	return ids
}

// GetAllForUser возвращает все URL пользователя в виде map[id]originalURL.
func (r *URLRepository) GetAllForUser(userID string) map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userStore, ok := r.store[userID]
	if !ok || len(userStore) == 0 {
		return map[string]string{}
	}

	result := make(map[string]string, len(userStore))
	for k, v := range userStore {
		result[k] = v
	}
	return result
}

// Get возвращает оригинальный URL по его ID.
//
// Если URL не найден, возвращает model.ErrNotFound.
func (r *URLRepository) Get(id string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, userStore := range r.store {
		if url, exists := userStore[id]; exists {
			return url, nil
		}
	}
	return "", model.ErrNotFound
}

// MarkAsDeleted помечает указанные URL пользователя как удалённые.
func (r *URLRepository) MarkAsDeleted(userID string, ids []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userStore, ok := r.store[userID]
	if !ok {
		return nil
	}

	for _, id := range ids {
		delete(userStore, id)
	}

	r.saveToFile()
	return nil
}

// Ping проверяет доступность хранилища. Всегда возвращает nil для in-memory реализации.
func (r *URLRepository) Ping(ctx context.Context) error {
	return nil
}

// GetStats возвращает количество URL и уникальных пользователей.
func (r *URLRepository) GetStats(ctx context.Context) (int, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var urls int
	for _, userStore := range r.store {
		urls += len(userStore)
	}
	return urls, len(r.store), nil
}

// saveToFile сохраняет данные на диск в JSON формате.
func (r *URLRepository) saveToFile() {
	if r.file == "" {
		return
	}
	data, err := json.MarshalIndent(r.store, "", " ")
	if err != nil {
		return
	}

	_ = os.WriteFile(r.file, data, 0644)
}

// loadFromFile загружает данные из JSON файла, если он существует.
func (r *URLRepository) loadFromFile() {
	if r.file == "" {
		return
	}

	data, err := os.ReadFile(r.file)
	if err != nil {
		return
	}

	var loaded map[string]map[string]string
	if err := json.Unmarshal(data, &loaded); err != nil {
		return
	}

	r.store = loaded
}

// generateID создаёт случайный 8-символьный идентификатор для короткого URL.
func generateID() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := make([]byte, 8)
	for i := range id {
		id[i] = letters[r.Intn(len(letters))]
	}
	return string(id)
}
