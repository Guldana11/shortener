package repository

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"sync"
	"time"
)

var ErrURLExists = errors.New("url already exists")

type URLRepository struct {
	store map[string]map[string]string
	mu    sync.RWMutex
	file  string
}

const globalUserID = "__global__"

func NewURLRepository(filePath string) *URLRepository {
	r := &URLRepository{
		store: make(map[string]map[string]string),
		file:  filePath,
	}
	r.loadFromFile()
	return r
}

func (r *URLRepository) Create(originalURL string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.store[globalUserID] == nil {
		r.store[globalUserID] = make(map[string]string)
	}

	for id, url := range r.store[globalUserID] {
		if url == originalURL {
			return id, ErrURLExists
		}
	}

	id := generateID()
	r.store[globalUserID][id] = originalURL
	r.saveToFile()
	return id, nil
}

func (r *URLRepository) CreateWithID(id, originalURL string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.store[globalUserID] == nil {
		r.store[globalUserID] = make(map[string]string)
	}

	r.store[globalUserID][id] = originalURL
	r.saveToFile()
}

func (r *URLRepository) BatchCreate(urls []string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.store[globalUserID] == nil {
		r.store[globalUserID] = make(map[string]string)
	}

	ids := make([]string, len(urls))
	for i, u := range urls {
		var id string
		for {
			id = generateID()
			if _, exists := r.store[globalUserID][id]; !exists {
				break
			}
		}
		r.store[globalUserID][id] = u
		ids[i] = id
	}

	r.saveToFile()
	return ids
}

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

func (r *URLRepository) Get(id string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, urls := range r.store {
		if url, ok := urls[id]; ok {
			return url, true
		}
	}
	return "", false
}

func (r *URLRepository) Ping(ctx context.Context) error {
	return nil
}

func (r *URLRepository) saveToFile() {
	if r.file == "" {
		return
	}
	data, err := json.MarshalIndent(r.store, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(r.file, data, 0644)
}

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

func generateID() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := make([]byte, 8)
	for i := range id {
		id[i] = letters[r.Intn(len(letters))]
	}
	return string(id)
}
