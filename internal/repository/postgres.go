package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(dsn string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id TEXT PRIMARY KEY,
		original TEXT NOT NULL
	);
	`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) Create(originalURL string) string {
	id := generateID()
	if _, err := r.db.Exec("INSERT INTO urls (id, original) VALUES ($1, $2)", id, originalURL); err != nil {
		panic(fmt.Sprintf("failed to insert URL: %v", err))
	}
	return id
}

func (r *PostgresRepository) CreateWithID(id, originalURL string) {
	if _, err := r.db.Exec("INSERT INTO urls (id, original) VALUES ($1, $2)", id, originalURL); err != nil {
		panic(fmt.Sprintf("failed to insert URL: %v", err))
	}
}

func (r *PostgresRepository) Get(id string) (string, bool) {
	var original string
	err := r.db.QueryRow("SELECT original FROM urls WHERE id=$1", id).Scan(&original)
	if err == sql.ErrNoRows {
		return "", false
	} else if err != nil {
		panic(fmt.Sprintf("failed to select URL: %v", err))
	}
	return original, true
}

func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
