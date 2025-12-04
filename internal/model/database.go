package model

type Storage struct {
	UUID        string `db:"user_id"`
	ShortURL    string `db:"short_url"`
	OriginalURL string `db:"original_url"`
	IsDeleted   bool   `db:"is_deleted"`
}
