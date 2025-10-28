package main

import (
	"github.com/Guldana11/shortener/internal/config"
	"github.com/Guldana11/shortener/internal/handler"
	"github.com/Guldana11/shortener/internal/middleware"
	"github.com/Guldana11/shortener/internal/repository"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := config.Init()

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.InfoLevel)

	repo := repository.NewURLRepository(cfg.FileStoragePath)
	h := handler.NewURLHandler(cfg.BaseURL, repo)

	r := setupRouter(h)

	log.Infof("Server is running on %s", cfg.Address)
	if err := r.Run(cfg.Address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(h *handler.URLHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.GzipMiddleware())

	r.POST("/", h.PostHandler)
	r.GET("/:id", h.GetHandler)
	r.POST("/api/shorten", h.ShortenHandler)

	return r
}
