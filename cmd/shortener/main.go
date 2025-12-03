package main

import (
	"context"
	"database/sql"

	"github.com/Guldana11/shortener/internal/config"
	"github.com/Guldana11/shortener/internal/handler"
	"github.com/Guldana11/shortener/internal/middleware"
	"github.com/Guldana11/shortener/internal/repository"
	"github.com/Guldana11/shortener/internal/worker"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Init()
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	var repo repository.Repository

	if cfg.DatabaseDSN != "" {
		pool, err := pgxpool.New(context.Background(), cfg.DatabaseDSN)
		if err != nil {
			logger.Fatal("failed to connect to database", zap.Error(err))
		}
		defer pool.Close()

		sqlDB, err := sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			logger.Fatal("failed to open sql.DB for migrations", zap.Error(err))
		}
		defer sqlDB.Close()

		driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
		if err != nil {
			logger.Fatal("failed to create migrate driver", zap.Error(err))
		}

		m, err := migrate.NewWithDatabaseInstance(
			"file://migrations",
			"postgres",
			driver,
		)
		if err != nil {
			logger.Fatal("failed to create migrate instance", zap.Error(err))
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			logger.Fatal("failed to run migrations", zap.Error(err))
		}

		repo = repository.NewPostgresRepository(pool)
		logger.Info("Using PostgreSQL storage")
	} else if cfg.FileStoragePath != "" {
		repo = repository.NewURLRepository(cfg.FileStoragePath)
		logger.Info("Using file storage", zap.String("file", cfg.FileStoragePath))
	} else {
		repo = repository.NewURLRepository("")
		logger.Info("Using in-memory storage")
	}

	deleteWorker := worker.NewDeleteWorker(repo, 1000)

	h := handler.NewURLHandler(cfg.BaseURL, repo, deleteWorker)

	r := setupRouter(h, logger)

	logger.Info("Server is starting...", zap.String("address", cfg.Address))
	if err := r.Run(cfg.Address); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func setupRouter(h *handler.URLHandler, logger *zap.Logger) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.GzipMiddleware())

	r.GET("/ping", h.PingHandler)

	r.POST("/", h.PostHandler)
	r.GET("/:id", h.GetHandler)
	r.POST("/api/shorten", h.ShortenHandler)
	r.POST("/api/shorten/batch", h.ShortenBatchHandler)
	r.GET("/api/user/urls", h.GetUserURLs)

	r.DELETE("/api/user/urls", h.DeleteUserURLs)

	return r
}
