// Package handler содержит HTTP-хендлеры сервиса сокращения URL.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Guldana11/shortener/internal/audit"
	"github.com/Guldana11/shortener/internal/model"
	"github.com/Guldana11/shortener/internal/repository"
	"github.com/Guldana11/shortener/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"go.uber.org/zap"
)

// Pinger описывает интерфейс проверки доступности хранилища.
type Pinger interface {
	// Ping проверяет доступность сервиса хранения.
	Ping(ctx context.Context) error
}

// DeleteWorkerInterface описывает очередь фонового удаления URL пользователя.
type DeleteWorkerInterface interface {
	// EnqueueDeletion добавляет список URL в очередь удаления.
	EnqueueDeletion(userID string, ids []string)
}

// URLHandler реализует HTTP-хендлеры сервиса сокращения URL.
type URLHandler struct {
	Repo         repository.Repository
	BaseURL      string
	logger       *zap.Logger
	DeleteWorker DeleteWorkerInterface
	Publisher    *audit.Publisher
}

// NewURLHandler создаёт новый экземпляр URLHandler.
func NewURLHandler(baseURL string, repo repository.Repository, dw DeleteWorkerInterface) *URLHandler {
	logger, _ := zap.NewProduction()
	return &URLHandler{
		Repo:         repo,
		BaseURL:      baseURL,
		logger:       logger,
		DeleteWorker: dw,
		Publisher:    publisher,
	}
}

// PostHandler обрабатывает POST /
// Принимает URL в теле запроса и возвращает сокращённый URL в виде строки.
func (h *URLHandler) PostHandler(c *gin.Context) {
	userID, err := service.ValidateUserCookie(c.Request)
	if err != nil || userID == "" {
		cookie := service.GenerateUserCookie()
		http.SetCookie(c.Writer, cookie)
		userID = cookie.Value[:36] // UUID без дефисов
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("failed to read request body", zap.Error(err))
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
	if len(body) == 0 {
		c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	originalURL := string(body)

	id, err := h.Repo.CreateForUser(userID, originalURL)
	if err != nil {
		if errors.Is(err, repository.ErrURLExists) {
			shortURL, _ := url.JoinPath(h.BaseURL, id)
			c.String(http.StatusConflict, shortURL)
			return
		}
		h.logger.Error("failed to create URL", zap.Error(err))
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	shortURL, err := url.JoinPath(h.BaseURL, id)
	if err != nil {
		h.logger.Error("failed to join URL path", zap.Error(err))
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	c.String(http.StatusCreated, shortURL)

	h.Publisher.Publish(audit.Event{
		TS:     time.Now().Unix(),
		Action: "shorten",
		UserID: userID,
		URL:    originalURL,
	})

}

// GetHandler обрабатывает GET /:id
// Выполняет редирект на оригинальный URL по его идентификатору.
func (h *URLHandler) GetHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" || containsSlash(id) {
		c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	original, err := h.Repo.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrNotFound):
			c.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		case errors.Is(err, model.ErrDeleted):
			c.String(http.StatusGone, http.StatusText(http.StatusGone))
		default:
			c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, original)

	userID, _ := service.ValidateUserCookie(c.Request)

	h.Publisher.Publish(audit.Event{
		TS:     time.Now().Unix(),
		Action: "follow",
		UserID: userID,
		URL:    original,
	})

}

// ShortenHandler обрабатывает POST /api/shorten
// Принимает JSON с полем "url" и возвращает JSON с короткой ссылкой.
func (h *URLHandler) ShortenHandler(c *gin.Context) {
	// Получаем userID
	userID, err := service.ValidateUserCookie(c.Request)
	if err != nil || userID == "" {
		cookie := service.GenerateUserCookie()
		http.SetCookie(c.Writer, cookie)
		userID = cookie.Value[:36]
	}

	var req model.ShortenRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil || req.URL == "" {
		h.logger.Error("failed to decode JSON or missing URL", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}

	id, err := h.Repo.CreateForUser(userID, req.URL)
	if err != nil {
		if errors.Is(err, repository.ErrURLExists) {
			shortURL, _ := url.JoinPath(h.BaseURL, id)
			c.JSON(http.StatusConflict, model.ShortenResponse{
				Result: shortURL,
			})
			return
		}

		h.logger.Error("failed to create URL", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	shortURL, err := url.JoinPath(h.BaseURL, id)
	if err != nil {
		h.logger.Error("failed to join URL path", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	c.JSON(http.StatusCreated, model.ShortenResponse{
		Result: shortURL,
	})

	h.Publisher.Publish(audit.Event{
		TS:     time.Now().Unix(),
		Action: "shorten",
		UserID: userID,
		URL:    req.URL,
	})

}

// PingHandler обрабатывает GET /ping
// Проверяет доступность базы данных.
func (h *URLHandler) PingHandler(c *gin.Context) {
	if h.Repo == nil {
		c.String(http.StatusInternalServerError, "database not configured")
		return
	}

	if err := h.Repo.Ping(c); err != nil {
		c.String(http.StatusInternalServerError, "database unreachable")
		return
	}
	c.String(http.StatusOK, "pong")
}

// ShortenBatchHandler обрабатывает POST /api/shorten/batch
// Позволяет создать несколько сокращённых URL за один запрос.
func (h *URLHandler) ShortenBatchHandler(c *gin.Context) {
	userID, err := service.ValidateUserCookie(c.Request)
	if err != nil || userID == "" {
		cookie := service.GenerateUserCookie()
		http.SetCookie(c.Writer, cookie)
		userID = cookie.Value[:36]
	}

	var req []struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil || len(req) == 0 {
		h.logger.Error("failed to decode batch request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	for i, item := range req {
		if item.OriginalURL == "" || item.CorrelationID == "" {
			h.logger.Warn("empty URL or correlation_id in batch", zap.Int("index", i))
			c.JSON(http.StatusBadRequest, gin.H{"error": "each item must have correlation_id and original_url"})
			return
		}
	}

	type responseItem struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}

	responses := make([]responseItem, len(req))

	for i, item := range req {
		id, err := h.Repo.CreateForUser(userID, item.OriginalURL)

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save URL"})
				return
			}
		}

		shortURL, _ := url.JoinPath(h.BaseURL, id)
		responses[i] = responseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		}
	}

	c.JSON(http.StatusCreated, responses)
}

// GetUserURLs обрабатывает GET /api/user/urls
// Возвращает список всех URL текущего пользователя.
func (h *URLHandler) GetUserURLs(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	userID, exists := c.Get("userID")
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	urls := h.Repo.GetAllForUser(userID.(string))
	if len(urls) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	type respPair struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	resp := make([]respPair, 0, len(urls))
	for id, original := range urls {
		shortURL, err := url.JoinPath(h.BaseURL, id)
		if err != nil {
			h.logger.Error("failed to build short URL", zap.Error(err), zap.String("id", id))
			c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}
		resp = append(resp, respPair{
			ShortURL:    shortURL,
			OriginalURL: original,
		})
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteUserURLs обрабатывает DELETE /api/user/urls
// Принимает список идентификаторов URL и отправляет их в очередь удаления.
func (h *URLHandler) DeleteUserURLs(c *gin.Context) {
	userID, err := service.ValidateUserCookie(c.Request)
	if err != nil || userID == "" {
		c.String(http.StatusUnauthorized, "unauthorized")
		return
	}

	var ids []string
	if err := json.NewDecoder(c.Request.Body).Decode(&ids); err != nil {
		c.String(http.StatusBadRequest, "invalid request body")
		return
	}

	h.DeleteWorker.EnqueueDeletion(userID, ids)

	c.Status(http.StatusAccepted)
}

func containsSlash(s string) bool {
	for _, c := range s {
		if c == '/' {
			return true
		}
	}
	return false
}
