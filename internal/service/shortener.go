package service

import (
	"errors"
	"net/url"
	"time"

	"github.com/Guldana11/shortener/internal/audit"
	"github.com/Guldana11/shortener/internal/model"
	"github.com/Guldana11/shortener/internal/repository"
	"go.uber.org/zap"
)

// URLPair represents a short URL and its original URL.
type URLPair struct {
	ShortURL    string
	OriginalURL string
}

// ShortenResult contains the result of a shorten operation.
type ShortenResult struct {
	ShortURL string
	Conflict bool // true if URL already existed
}

var (
	ErrNotFound = model.ErrNotFound
	ErrDeleted  = model.ErrDeleted
)

// URLService contains the shared business logic for URL shortening.
type URLService struct {
	Repo      repository.Repository
	BaseURL   string
	Publisher *audit.Publisher
	Logger    *zap.Logger
}

// ShortenURL creates a short URL for the given user and original URL.
// Returns the full short URL and whether a conflict (duplicate) occurred.
func (s *URLService) ShortenURL(userID, originalURL string) (ShortenResult, error) {
	id, err := s.Repo.CreateForUser(userID, originalURL)
	if err != nil {
		if errors.Is(err, repository.ErrURLExists) {
			shortURL, _ := url.JoinPath(s.BaseURL, id)
			return ShortenResult{ShortURL: shortURL, Conflict: true}, nil
		}
		s.Logger.Error("failed to create URL", zap.Error(err))
		return ShortenResult{}, err
	}

	shortURL, err := url.JoinPath(s.BaseURL, id)
	if err != nil {
		s.Logger.Error("failed to build short URL", zap.Error(err))
		return ShortenResult{}, err
	}

	s.Publisher.Publish(audit.Event{
		TS:     time.Now().Unix(),
		Action: "shorten",
		UserID: userID,
		URL:    originalURL,
	})

	return ShortenResult{ShortURL: shortURL}, nil
}

// ExpandURL returns the original URL for the given short ID.
// Returns model.ErrNotFound or model.ErrDeleted if the URL doesn't exist.
func (s *URLService) ExpandURL(id string) (string, error) {
	original, err := s.Repo.Get(id)
	if err != nil {
		if !errors.Is(err, model.ErrNotFound) && !errors.Is(err, model.ErrDeleted) {
			s.Logger.Error("failed to get URL", zap.Error(err))
		}
		return "", err
	}
	return original, nil
}

// ListUserURLs returns all URLs for the given user.
func (s *URLService) ListUserURLs(userID string) []URLPair {
	urls := s.Repo.GetAllForUser(userID)
	result := make([]URLPair, 0, len(urls))
	for id, original := range urls {
		shortURL, _ := url.JoinPath(s.BaseURL, id)
		result = append(result, URLPair{
			ShortURL:    shortURL,
			OriginalURL: original,
		})
	}
	return result
}
