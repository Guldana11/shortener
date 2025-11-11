package handler

import (
	"context"
	"encoding/json"
	"github.com/Guldana11/shortener/internal/model"
	"github.com/Guldana11/shortener/internal/repository"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type URLHandler struct {
	Repo    repository.Repository
	BaseURL string
	Logger  *zap.Logger
	DB      Pinger
}

func NewURLHandler(baseURL string, repo repository.Repository, db Pinger, logger *zap.Logger) *URLHandler {
	return &URLHandler{
		Repo:    repo,
		BaseURL: baseURL,
		Logger:  logger,
		DB:      db,
	}
}

func (h *URLHandler) PostHandler(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	if len(body) == 0 {
		c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}
	originalURL := string(body)
	id := h.Repo.Create(originalURL)
	shortURL, _ := url.JoinPath(h.BaseURL, id)
	c.String(http.StatusCreated, shortURL)
}

func (h *URLHandler) GetHandler(c *gin.Context) {
	id := c.Param("id")
	original, ok := h.Repo.Get(id)
	if !ok {
		c.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, original)
}

func (h *URLHandler) ShortenHandler(c *gin.Context) {
	var req model.ShortenRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil || req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	id := h.Repo.Create(req.URL)
	shortURL, _ := url.JoinPath(h.BaseURL, id)
	c.JSON(http.StatusCreated, model.ShortenResponse{Result: shortURL})
}

func (h *URLHandler) PingHandler(c *gin.Context) {
	if err := h.DB.Ping(c.Request.Context()); err != nil {
		c.String(http.StatusInternalServerError, "DB not available")
		return
	}
	c.String(http.StatusOK, "pong")
}
