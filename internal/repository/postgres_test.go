package repository

import (
	"context"
	"os"
	"testing"

	"github.com/Guldana11/shortener/internal/model"
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

	_, err = db.Exec(context.Background(), `
	CREATE TABLE urls (
		id TEXT PRIMARY KEY,
		original_url TEXT NOT NULL,
		user_id TEXT,
		is_deleted BOOLEAN DEFAULT FALSE
	);
	`)
	if err != nil {
		t.Fatalf("Не удалось создать таблицу urls: %v", err)
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
			id, err := repo.Create(tt.originalURL)
			if err != nil {
				t.Fatalf("Create вернул ошибку: %v", err)
			}
			if id == "" {
				t.Fatal("Create вернул пустой ID")
			}

			got, err := repo.Get(id)
			if err != nil {
				t.Fatalf("Get вернул ошибку: %v", err)
			}
			if got != tt.originalURL {
				t.Fatalf("ожидали %v, получили %v", tt.originalURL, got)
			}
		})
	}
}

func TestPostgresRepository_CreateWithID(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)

	id := "customID"
	url := "https://example.net"
	repo.CreateWithID(id, url)

	got, err := repo.Get(id)
	if err != nil {
		t.Fatalf("Get вернул ошибку: %v", err)
	}
	if got != url {
		t.Errorf("ожидали %v, получили %v", url, got)
	}
}

func TestPostgresRepository_CreateForUserAndGetAll(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	repo := NewPostgresRepository(db)

	userID := "user1"
	urls := []string{"https://a.com", "https://b.com"}

	for _, u := range urls {
		_, err := repo.CreateForUser(userID, u)
		if err != nil {
			t.Fatalf("CreateForUser вернул ошибку: %v", err)
		}
	}

	all := repo.GetAllForUser(userID)
	if len(all) != len(urls) {
		t.Fatalf("ожидали %d URL, получили %d", len(urls), len(all))
	}

	for _, u := range urls {
		found := false
		for _, v := range all {
			if v == u {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("URL %s не найден в GetAllForUser", u)
		}
	}
}

func TestPostgresRepository_Ping(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)

	if err := repo.Ping(context.Background()); err != nil {
		t.Errorf("Ping() вернул ошибку: %v", err)
	}
}

func TestPostgresRepository_MarkAsDeleted(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	userID := "user123"

	id1, _ := repo.CreateForUser(userID, "https://example.com/1")
	id2, _ := repo.CreateForUser(userID, "https://example.com/2")

	tests := []struct {
		name   string
		ids    []string
		userID string
	}{
		{
			name:   "Помечаем два URL как удаленные",
			ids:    []string{id1, id2},
			userID: userID,
		},
		{
			name:   "Пустой список ID не вызывает ошибки",
			ids:    []string{},
			userID: userID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.MarkAsDeleted(tt.userID, tt.ids)
			if err != nil {
				t.Fatalf("MarkAsDeleted вернул ошибку: %v", err)
			}

			for _, id := range tt.ids {
				_, err := repo.Get(id)
				if err != model.ErrDeleted {
					t.Errorf("URL с ID %s должен быть помечен как удалённый, получили: %v", id, err)
				}
			}
		})
	}
}
