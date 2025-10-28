package repository

import (
	"encoding/json"
	"math/rand"
	"os"
	"sync"
	"time"
)

type URLRepository struct {
	store map[string]string
	mu    sync.RWMutex
	file  string
}

func NewURLRepository(filePath string) *URLRepository {
	r := &URLRepository{
		store: make(map[string]string),
		file:  filePath,
	}

	r.loadFromFile()
	return r
}

func (r *URLRepository) Create(originalURL string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := generateID()
	r.store[id] = originalURL

	r.saveToFile()
	return id
}

func (r *URLRepository) CreateWithID(id, originalURL string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[id] = originalURL
}

func (r *URLRepository) Get(id string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	url, ok := r.store[id]
	return url, ok
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

	var loaded map[string]string
	if err := json.Unmarshal(data, &loaded); err != nil {
		return
	}

	for id, url := range loaded {
		r.CreateWithID(id, url)
	}
}

func generateID() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())

	id := make([]byte, 8)
	for i := range id {
		id[i] = letters[rand.Intn(len(letters))]
	}
	return string(id)
}
