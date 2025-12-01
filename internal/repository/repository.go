package repository

import "context"

type Repository interface {
	Create(originalURL string) (string, error)
	CreateWithID(id, originalURL string)
	Get(id string) (string, bool, bool)
	Ping(ctx context.Context) error
	CreateForUser(userID, originalURL string) (string, error)
	GetAllForUser(userID string) map[string]string
	MarkAsDeleted(userID string, ids []string) error
}
