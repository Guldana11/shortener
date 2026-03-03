package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/Guldana11/shortener/internal/audit"
	"github.com/Guldana11/shortener/internal/model"
	"github.com/Guldana11/shortener/internal/repository"
	"github.com/Guldana11/shortener/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type mockRepo struct {
	store map[string]map[string]string
	err   error
}

func newMockRepo(err error) *mockRepo {
	return &mockRepo{
		store: make(map[string]map[string]string),
		err:   err,
	}
}

func (m *mockRepo) Ping(ctx context.Context) error {
	return m.err
}

func (m *mockRepo) Create(originalURL string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if m.store["__global__"] == nil {
		m.store["__global__"] = make(map[string]string)
	}
	id := "global" + strconv.Itoa(len(m.store["__global__"])+1)
	m.store["__global__"][id] = originalURL
	return id, nil
}

func (m *mockRepo) CreateWithID(id string, originalURL string) {
	if m.store["__global__"] == nil {
		m.store["__global__"] = make(map[string]string)
	}
	m.store["__global__"][id] = originalURL
}

func (m *mockRepo) CreateForUser(userID, originalURL string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if m.store[userID] == nil {
		m.store[userID] = make(map[string]string)
	}
	id := "id" + strconv.Itoa(len(m.store[userID])+1)
	m.store[userID][id] = originalURL
	return id, nil
}

func (m *mockRepo) Get(id string) (string, error) {
	for _, urls := range m.store {
		if val, ok := urls[id]; ok {
			return val, nil
		}
	}
	return "", model.ErrNotFound
}

func (m *mockRepo) GetAllForUser(userID string) map[string]string {
	if m.store[userID] == nil {
		return nil
	}
	return m.store[userID]
}

func (m *mockRepo) MarkAsDeleted(userID string, ids []string) error {
	return nil
}

func (m *mockRepo) GetStats(ctx context.Context) (int, int, error) {
	var urls int
	for _, userStore := range m.store {
		urls += len(userStore)
	}
	return urls, len(m.store), nil
}

type mockDeleteWorker struct{}

func (m *mockDeleteWorker) EnqueueDeletion(userID string, ids []string) {}

type mockPublisher struct{}

func (m *mockPublisher) Publish(data interface{}) error {
	return nil
}

func TestPostHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockRepo(nil)
	pub := &audit.Publisher{}
	h := &URLHandler{
		Repo:         repo,
		BaseURL:      "http://localhost:8080",
		logger:       zap.NewNop(),
		DeleteWorker: &mockDeleteWorker{},
		Publisher:    pub,
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://example.com"))

	h.PostHandler(c)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected %d, got %d", http.StatusCreated, rec.Code)
	}

	body, _ := io.ReadAll(rec.Body)
	if !strings.HasPrefix(string(body), "http://localhost:8080/") {
		t.Errorf("unexpected body: %s", string(body))
	}
}

func TestGetHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockRepo(nil)
	id, _ := repo.CreateForUser("user1", "https://example.com")
	pub := &audit.Publisher{}
	h := &URLHandler{
		Repo:      repo,
		BaseURL:   "http://localhost:8080",
		logger:    zap.NewNop(),
		Publisher: pub,
	}

	rec := httptest.NewRecorder()
	c, r := gin.CreateTestContext(rec)
	r.GET("/:id", h.GetHandler)
	req := httptest.NewRequest(http.MethodGet, "/"+id, nil)
	c.Request = req
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected %d, got %d", http.StatusTemporaryRedirect, rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "https://example.com" {
		t.Errorf("expected Location header 'https://example.com', got %s", loc)
	}
}

func TestShortenHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockRepo(nil)
	pub := &audit.Publisher{}
	h := &URLHandler{
		Repo:         repo,
		BaseURL:      "http://localhost:8080",
		logger:       zap.NewNop(),
		DeleteWorker: &mockDeleteWorker{},
		Publisher:    pub,
	}

	body := `{"url":"https://example.com"}`
	rec := httptest.NewRecorder()
	router := gin.Default()
	router.POST("/api/shorten", h.ShortenHandler)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected %d, got %d", http.StatusCreated, rec.Code)
	}
	var resp model.ShortenResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if !strings.HasPrefix(resp.Result, "http://localhost:8080/") {
		t.Errorf("unexpected short URL: %s", resp.Result)
	}
}

func TestPingHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		repo   repository.Repository
		status int
		body   string
	}{
		{nil, http.StatusInternalServerError, "database not configured"},
		{newMockRepo(errors.New("fail")), http.StatusInternalServerError, "database unreachable"},
		{newMockRepo(nil), http.StatusOK, "pong"},
	}

	for _, tt := range tests {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "/ping", nil)
		h := &URLHandler{Repo: tt.repo, logger: zap.NewNop()}
		h.PingHandler(c)

		if rec.Code != tt.status {
			t.Errorf("expected %d, got %d", tt.status, rec.Code)
		}
		body, _ := io.ReadAll(rec.Body)
		if string(body) != tt.body {
			t.Errorf("expected body '%s', got '%s'", tt.body, string(body))
		}
	}
}

func TestDeleteUserURLs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockRepo(nil)
	h := &URLHandler{
		Repo:         repo,
		BaseURL:      "http://localhost:8080",
		logger:       zap.NewNop(),
		DeleteWorker: &mockDeleteWorker{},
	}

	cookie := service.GenerateUserCookie()
	userID, _ := service.ValidateUserCookie(&http.Request{
		Header: http.Header{"Cookie": []string{cookie.String()}},
	})
	id1, _ := repo.CreateForUser(userID, "https://example.com")

	rec := httptest.NewRecorder()
	c, r := gin.CreateTestContext(rec)
	r.DELETE("/api/user/urls", h.DeleteUserURLs)
	body := `["` + id1 + `"]`
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBufferString(body))
	req.AddCookie(cookie)
	c.Request = req
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("expected %d, got %d", http.StatusAccepted, rec.Code)
	}
}
