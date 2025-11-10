package main

import (
	"context"

	"github.com/Guldana11/shortener/internal/config"
	"github.com/Guldana11/shortener/internal/handler"
	"github.com/Guldana11/shortener/internal/middleware"
	"github.com/Guldana11/shortener/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Init()

	logger, err := zap.NewProduction()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	repo := repository.NewURLRepository(cfg.FileStoragePath)

	var db *pgxpool.Pool
	if cfg.DatabaseDSN != "" {
		db, err = pgxpool.New(context.Background(), cfg.DatabaseDSN)
		if err != nil {
			logger.Fatal("failed to connect to database", zap.Error(err))
		}
		logger.Info("Connected to PostgreSQL")
	}

	h := handler.NewURLHandler(cfg.BaseURL, repo, db)

	r := setupRouter(h, logger)

	logger.Info("Server is starting...",
		zap.String("address", cfg.Address),
		zap.String("baseURL", cfg.BaseURL),
		zap.String("storageFile", cfg.FileStoragePath),
		zap.String("databaseDSN", cfg.DatabaseDSN),
	)

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

	return r
}
