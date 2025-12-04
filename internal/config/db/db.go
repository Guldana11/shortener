package db

import (
	"context"
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(ctx context.Context, dsn string, migrationsPath string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		pool.Close()
		return nil, err
	}
	defer sqlDB.Close()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		pool.Close()
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		pool.Close()
		return nil, err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
