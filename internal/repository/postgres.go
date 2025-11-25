package repository

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id TEXT PRIMARY KEY,
		original_url TEXT NOT NULL,
		user_id TEXT,
		UNIQUE(original_url, user_id)
	);`
	_, err := db.Exec(context.Background(), query)
	if err != nil {
		log.Fatal("failed to create table:", err)
	}
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(originalURL string) (string, error) {
	return r.CreateForUser("", originalURL)
}

func (r *PostgresRepository) CreateWithID(id, originalURL string) {
	_, err := r.db.Exec(context.Background(),
		"INSERT INTO urls (id, original_url) VALUES ($1, $2)",
		id, originalURL,
	)
	if err != nil {
		log.Println("failed to insert URL:", err)
	}
}

func (r *PostgresRepository) CreateForUser(userID, originalURL string) (string, error) {
	id := generateID()

	_, err := r.db.Exec(context.Background(),
		"INSERT INTO urls (id, original_url, user_id) VALUES ($1, $2, $3)",
		id, originalURL, userID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			existingID, ok := r.getIDByOriginalForUser(userID, originalURL)
			if ok {
				return existingID, ErrURLExists
			}
			return "", ErrURLExists
		}
		return "", err
	}
	return id, nil
}

func (r *PostgresRepository) getIDByOriginalForUser(userID, originalURL string) (string, bool) {
	var id string
	err := r.db.QueryRow(context.Background(),
		"SELECT id FROM urls WHERE original_url=$1 AND user_id=$2",
		originalURL, userID,
	).Scan(&id)
	if err != nil {
		return "", false
	}
	return id, true
}

func (r *PostgresRepository) GetAllForUser(userID string) map[string]string {
	rows, err := r.db.Query(context.Background(),
		"SELECT id, original_url FROM urls WHERE user_id=$1",
		userID,
	)
	if err != nil {
		log.Println("failed to query URLs for user:", err)
		return nil
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var id, url string
		if err := rows.Scan(&id, &url); err == nil {
			result[id] = url
		}
	}
	return result
}

func (r *PostgresRepository) Get(id string) (string, bool) {
	var original string
	err := r.db.QueryRow(context.Background(),
		"SELECT original_url FROM urls WHERE id=$1",
		id,
	).Scan(&original)
	if err != nil {
		return "", false
	}
	return original, true
}

func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}
