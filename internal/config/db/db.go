// Package db предоставляет функции для работы с базой данных PostgreSQL,
// включая создание пула подключений и применение миграций.
package db

import (
	"context"
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool создаёт пул подключений к базе данных PostgreSQL и применяет миграции.
//
// ctx — контекст для управления временем жизни подключений.
// dsn — строка подключения к базе данных PostgreSQL.
// migrationsPath — путь к папке с миграциями.
//
// Возвращает пул подключений *pgxpool.Pool и ошибку в случае неудачи.
//
// Функция выполняет следующие шаги:
// 1. Создаёт пул подключений к базе PostgreSQL через pgxpool.
// 2. Создаёт экземпляр sql.DB для драйвера миграций.
// 3. Создаёт драйвер для миграций с использованием sql.DB.
// 4. Применяет все миграции из migrationsPath.
// 5. Если миграции прошли успешно или изменений нет, возвращает пул.
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
