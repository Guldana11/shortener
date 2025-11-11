package repository

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id TEXT PRIMARY KEY,
		original_url TEXT NOT NULL
	);`
	_, err := db.Exec(context.Background(), query)
	if err != nil {
		log.Fatal("failed to create table:", err)
	}
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(originalURL string) string {
	id := generateID()
	_, err := r.db.Exec(context.Background(), "INSERT INTO urls (id, original_url) VALUES ($1, $2)", id, originalURL)
	if err != nil {
		log.Println("failed to insert URL:", err)
		return ""
	}
	return id
}

func (r *PostgresRepository) CreateWithID(id, originalURL string) {
	_, err := r.db.Exec(context.Background(), "INSERT INTO urls (id, original_url) VALUES ($1, $2)", id, originalURL)
	if err != nil {
		log.Println("failed to insert URL:", err)
	}
}

func (r *PostgresRepository) Get(id string) (string, bool) {
	var original string
	err := r.db.QueryRow(context.Background(), "SELECT original_url FROM urls WHERE id=$1", id).Scan(&original)
	if err != nil {
		return "", false
	}
	return original, true
}

func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}
