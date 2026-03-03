package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/Guldana11/shortener/internal/model"
	"github.com/Guldana11/shortener/internal/repository"
	"github.com/Guldana11/shortener/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"go.uber.org/zap"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type DeleteWorkerInterface interface {
	EnqueueDeletion(userID string, ids []string)
}

type URLHandler struct {
	Repo          repository.Repository
	BaseURL       string
	logger        *zap.Logger
	DeleteWorker  DeleteWorkerInterface
	Svc           *service.URLService
	TrustedSubnet *net.IPNet
}

func NewURLHandler(baseURL string, repo repository.Repository, dw DeleteWorkerInterface, svc *service.URLService, trustedSubnet string) *URLHandler {
	logger, _ := zap.NewProduction()
	h := &URLHandler{
		Repo:         repo,
		BaseURL:      baseURL,
		logger:       logger,
		DeleteWorker: dw,
		Svc:          svc,
	}
	if trustedSubnet != "" {
		_, ipNet, err := net.ParseCIDR(trustedSubnet)
		if err == nil {
			h.TrustedSubnet = ipNet
		}
	}
	return h
}

// POST /
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

	result, err := h.Svc.ShortenURL(userID, string(body))
	if err != nil {
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
	if result.Conflict {
		c.String(http.StatusConflict, result.ShortURL)
		return
	}

	c.String(http.StatusCreated, result.ShortURL)
}

// GET /:id
func (h *URLHandler) GetHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" || containsSlash(id) {
		c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	original, err := h.Svc.ExpandURL(id)
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
}

// POST /api/shorten
func (h *URLHandler) ShortenHandler(c *gin.Context) {
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

	result, err := h.Svc.ShortenURL(userID, req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}
	if result.Conflict {
		c.JSON(http.StatusConflict, model.ShortenResponse{Result: result.ShortURL})
		return
	}

	c.JSON(http.StatusCreated, model.ShortenResponse{Result: result.ShortURL})
}

// GET /ping
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

// POST /api/shorten/batch
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

// GET /api/user/urls
func (h *URLHandler) GetUserURLs(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	userID, exists := c.Get("userID")
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	urls := h.Svc.ListUserURLs(userID.(string))
	if len(urls) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	type respPair struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	resp := make([]respPair, 0, len(urls))
	for _, u := range urls {
		resp = append(resp, respPair{
			ShortURL:    u.ShortURL,
			OriginalURL: u.OriginalURL,
		})
	}

	c.JSON(http.StatusOK, resp)
}

// DELETE /api/user/urls
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

// GET /api/internal/stats
func (h *URLHandler) StatsHandler(c *gin.Context) {
	if h.TrustedSubnet == nil {
		c.String(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	ip := net.ParseIP(c.GetHeader("X-Real-IP"))
	if ip == nil || !h.TrustedSubnet.Contains(ip) {
		c.String(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	urls, users, err := h.Repo.GetStats(c)
	if err != nil {
		h.logger.Error("failed to get stats", zap.Error(err))
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"urls":  urls,
		"users": users,
	})
}

func containsSlash(s string) bool {
	for _, c := range s {
		if c == '/' {
			return true
		}
	}
	return false
}
