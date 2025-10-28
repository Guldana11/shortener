package repository

type Repository interface {
	Create(originalURL string) string
	CreateWithID(id, originalURL string)
	Get(id string) (string, bool)
}
