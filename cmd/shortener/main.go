package main

import (
	"log"

	"github.com/Guldana11/shortener/internal/config"
	"github.com/Guldana11/shortener/internal/handler"
	"github.com/Guldana11/shortener/internal/repository"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Init()

	logger, err := zap.NewProduction()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	var repo repository.Repository
	var db handler.Pinger

	if cfg.DatabaseDSN != "" {
		pgRepo, err := repository.NewPostgresRepository(cfg.DatabaseDSN)
		if err == nil {
			repo = pgRepo
			db = pgRepo
		} else {
			log.Printf("failed to connect to DB: %v, fallback to file/memory", err)
		}
	}

	if repo == nil && cfg.FileStoragePath != "" {
		repo = repository.NewURLRepository(cfg.FileStoragePath)
	}

	if repo == nil {
		repo = repository.NewURLRepository("")
	}

	h := handler.NewURLHandler(cfg.BaseURL, repo, db)

	r := setupRouter(h)

	logger.Info("Server is starting...",
		zap.String("address", cfg.Address),
		zap.String("baseURL", cfg.BaseURL),
		zap.String("storageFile", cfg.FileStoragePath),
	)

	if err := r.Run(cfg.Address); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func setupRouter(h *handler.URLHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.POST("/", h.PostHandler)
	r.GET("/:id", h.GetHandler)
	r.POST("/api/shorten", h.ShortenHandler)
	r.GET("/ping", h.PingHandler)

	return r
}
