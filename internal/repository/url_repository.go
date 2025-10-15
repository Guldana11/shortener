package repository

import (
	"math/rand"
	"sync"
)

type URLRepository struct {
	store map[string]string
	mu    sync.RWMutex
}

func NewURLRepository() *URLRepository {
	return &URLRepository{
		store: make(map[string]string),
	}
}

func (r *URLRepository) CreateWithID(id, originalURL string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[id] = originalURL
}

func (r *URLRepository) Create(originalURL string) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	id := string(b)

	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[id] = originalURL

	return id
}

func (r *URLRepository) Get(id string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	url, ok := r.store[id]
	return url, ok
}
