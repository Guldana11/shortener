package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Guldana11/shortener/internal/model"
	"github.com/Guldana11/shortener/internal/repository"

	"github.com/gin-gonic/gin"
)

type URLHandler struct {
	repo    *repository.URLRepository
	BaseURL string
}

func NewURLHandler(baseURL string, repo *repository.URLRepository) *URLHandler {
	return &URLHandler{
		repo:    repo,
		BaseURL: baseURL,
	}
}

func (h *URLHandler) PostHandler(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 {
		c.String(http.StatusBadRequest, "bad request")
		return
	}

	id := h.repo.Create(string(body))

	baseURL := h.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	shortURL := fmt.Sprintf("%s/%s", baseURL, id)
	c.String(http.StatusCreated, shortURL)
}

func (h *URLHandler) GetHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" || containsSlash(id) {
		c.String(http.StatusBadRequest, "bad request")
		return
	}

	original, ok := h.repo.Get(id)
	if !ok {
		c.String(http.StatusNotFound, "not found")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, original)
}

func (h *URLHandler) ShortenHandler(c *gin.Context) {
	var req model.ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	shortID := h.repo.Create(req.URL)
	shortURL := h.BaseURL + "/" + shortID

	c.JSON(http.StatusCreated, model.ShortenResponse{Result: shortURL})
}

func containsSlash(s string) bool {
	for _, c := range s {
		if c == '/' {
			return true
		}
	}
	return false
}
