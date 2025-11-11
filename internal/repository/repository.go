package repository

import "context"

type Repository interface {
	Create(originalURL string) string
	CreateWithID(id, originalURL string)
	Get(id string) (string, bool)
	Ping(ctx context.Context) error
}
