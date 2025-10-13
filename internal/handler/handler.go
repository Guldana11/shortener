package handler

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
)

type URLHandler struct {
	store   map[string]string
	mu      sync.Mutex
	BaseURL string
}

func NewURLHandler(baseURL string) *URLHandler {
	return &URLHandler{
		store:   make(map[string]string),
		BaseURL: baseURL,
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

	shortURL := fmt.Sprintf("%s/%s", h.BaseURL, id)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(shortURL))
}

func (h *URLHandler) GetHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" && len(r.URL.Path) > 1 {
		id = r.URL.Path[1:]
	}

	if id == "" || containsSlash(id) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

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

func containsSlash(s string) bool {
	for _, c := range s {
		if c == '/' {
			return true
		}
	}
	return false
}
