package handler

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
)

type URLHandler struct {
	store map[string]string
	mu    sync.Mutex
}

func NewURLHandler() *URLHandler {
	return &URLHandler{
		store: make(map[string]string),
	}
}

func (h *URLHandler) PostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	id := generateID()
	h.mu.Lock()
	h.store[id] = string(body)
	h.mu.Unlock()

	shortURL := fmt.Sprintf("http://localhost:8080/%s", id)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(shortURL))
}

func (h *URLHandler) GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	id := r.URL.Path[1:]
	h.mu.Lock()
	original, ok := h.store[id]
	h.mu.Unlock()
	if !ok {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", original)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func generateID() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
