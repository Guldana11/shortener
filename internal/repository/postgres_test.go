package repository

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func getTestDB(t *testing.T) *pgxpool.Pool {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN не задана")
	}

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("Не удалось подключиться к БД: %v", err)
	}

	// Сбрасываем таблицу urls перед тестами
	_, err = db.Exec(context.Background(), "DROP TABLE IF EXISTS urls")
	if err != nil {
		t.Fatalf("Не удалось сбросить таблицу: %v", err)
	}

	return db
}

func TestNewPostgresRepository(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	got := NewPostgresRepository(db)
	if got == nil {
		t.Fatal("NewPostgresRepository вернул nil")
	}
}

func TestPostgresRepository_CreateAndGet(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)

	tests := []struct {
		name        string
		originalURL string
	}{
		{"Создание URL 1", "https://example.com"},
		{"Создание URL 2", "https://example.org"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := repo.Create(tt.originalURL)
			if id == "" {
				t.Fatal("Create вернул пустой ID")
			}

			got, ok := repo.Get(id)
			if !ok {
				t.Fatal("Get не вернул URL")
			}
			if got != tt.originalURL {
				t.Errorf("ожидали %v, получили %v", tt.originalURL, got)
			}
		})
	}
}

func TestPostgresRepository_CreateWithID(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)

	tests := []struct {
		name        string
		id          string
		originalURL string
	}{
		{"Создание с заданным ID", "customID", "https://example.net"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.CreateWithID(tt.id, tt.originalURL)

			got, ok := repo.Get(tt.id)
			if !ok {
				t.Fatal("Get не вернул URL после CreateWithID")
			}
			if got != tt.originalURL {
				t.Errorf("ожидали %v, получили %v", tt.originalURL, got)
			}
		})
	}
}

func TestPostgresRepository_Ping(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{"Ping успешен", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := repo.Ping(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
