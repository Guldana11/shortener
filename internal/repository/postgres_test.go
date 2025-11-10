package repository

import (
	"context"
	"database/sql"
	"os"
	"reflect"
	"testing"
)

func TestNewPostgresRepository(t *testing.T) {
	type args struct {
		dsn string
	}

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set")
	}

	tests := []struct {
		name    string
		args    args
		wantNil bool
		wantErr bool
	}{
		{
			name:    "valid DSN",
			args:    args{dsn: dsn},
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "invalid DSN",
			args:    args{dsn: "invalid_dsn"},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPostgresRepository(tt.args.dsn)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPostgresRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got == nil) != tt.wantNil {
				t.Errorf("NewPostgresRepository() got = %v, wantNil %v", got, tt.wantNil)
			}

			if got != nil && !reflect.TypeOf(got).AssignableTo(reflect.TypeOf(&PostgresRepository{})) {
				t.Errorf("got type %T, want *PostgresRepository", got)
			}
		})
	}
}

func TestPostgresRepository_Create(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set")
	}

	repo, err := NewPostgresRepository(dsn)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	type args struct {
		originalURL string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "insert URL 1",
			args: args{originalURL: "https://example.com/1"},
		},
		{
			name: "insert URL 2",
			args: args{originalURL: "https://example.com/2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := repo.Create(tt.args.originalURL)
			if id == "" {
				t.Errorf("Create() returned empty id")
			}

			got, ok := repo.Get(id)
			if !ok {
				t.Errorf("Create() did not save URL, id = %s", id)
			}
			if got != tt.args.originalURL {
				t.Errorf("Create() saved URL = %q, want %q", got, tt.args.originalURL)
			}
		})
	}
}

func TestPostgresRepository_CreateWithID(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set")
	}

	repo, err := NewPostgresRepository(dsn)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	type args struct {
		id          string
		originalURL string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "insert with custom ID 1",
			args: args{id: "custom1", originalURL: "https://example.com/1"},
		},
		{
			name: "insert with custom ID 2",
			args: args{id: "custom2", originalURL: "https://example.com/2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.CreateWithID(tt.args.id, tt.args.originalURL)

			got, ok := repo.Get(tt.args.id)
			if !ok {
				t.Errorf("CreateWithID() did not save URL for id = %s", tt.args.id)
			}
			if got != tt.args.originalURL {
				t.Errorf("CreateWithID() saved URL = %q, want %q", got, tt.args.originalURL)
			}
		})
	}
}

func TestPostgresRepository_Get(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set")
	}

	repo, err := NewPostgresRepository(dsn)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	repo.CreateWithID("id1", "https://example.com/1")
	repo.CreateWithID("id2", "https://example.com/2")

	type args struct {
		id string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{
			name:  "existing ID 1",
			args:  args{id: "id1"},
			want:  "https://example.com/1",
			want1: true,
		},
		{
			name:  "existing ID 2",
			args:  args{id: "id2"},
			want:  "https://example.com/2",
			want1: true,
		},
		{
			name:  "non-existent ID",
			args:  args{id: "unknown"},
			want:  "",
			want1: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := repo.Get(tt.args.id)
			if got != tt.want {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Get() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestPostgresRepository_Ping(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		db      *sql.DB
		args    args
		wantErr bool
	}{
		{
			name:    "Ping successful",
			db:      db,
			args:    args{ctx: context.Background()},
			wantErr: false,
		},
		{
			name:    "Ping with closed DB",
			db:      func() *sql.DB { d, _ := sql.Open("postgres", dsn); d.Close(); return d }(),
			args:    args{ctx: context.Background()},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &PostgresRepository{
				db: tt.db,
			}
			if err := r.Ping(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
