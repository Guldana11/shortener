package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/Guldana11/shortener/internal/audit"
	"github.com/Guldana11/shortener/internal/config"
	"github.com/Guldana11/shortener/internal/config/db"
	"github.com/Guldana11/shortener/internal/handler"
	"github.com/Guldana11/shortener/internal/middleware"
	"github.com/Guldana11/shortener/internal/repository"
	"github.com/Guldana11/shortener/internal/worker"
	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	printBuildInfo()

	go func() {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			zap.L().Fatal("pprof server failed", zap.Error(err))
		}
	}()

	cfg := config.Init()
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	var repo repository.Repository

	if cfg.DatabaseDSN != "" {
		pool, err := db.NewPostgresPool(context.Background(), cfg.DatabaseDSN, "file://migrations")
		if err != nil {
			logger.Fatal("failed to connect to database", zap.Error(err))
		}
		defer pool.Close()

		repo = repository.NewPostgresRepository(pool)
		logger.Info("Using PostgreSQL storage")
	} else if cfg.FileStoragePath != "" {
		repo = repository.NewURLRepository(cfg.FileStoragePath)
		logger.Info("Using file storage", zap.String("file", cfg.FileStoragePath))
	} else {
		repo = repository.NewURLRepository("")
		logger.Info("Using in-memory storage")
	}

	publisher := audit.NewPublisher()

	if cfg.AuditFile != "" {
		publisher.Subscribe(audit.NewFileObserver(cfg.AuditFile))
	}

	if cfg.AuditURL != "" {
		publisher.Subscribe(audit.NewHTTPObserver(cfg.AuditURL))
	}

	deleteWorker := worker.NewDeleteWorker(repo, 1000)
	h := handler.NewURLHandler(cfg.BaseURL, repo, deleteWorker, publisher)

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
	r.Use(middleware.UserCookieMiddleware())

	r.GET("/ping", h.PingHandler)

	r.POST("/", h.PostHandler)
	r.GET("/:id", h.GetHandler)
	r.POST("/api/shorten", h.ShortenHandler)
	r.POST("/api/shorten/batch", h.ShortenBatchHandler)
	r.GET("/api/user/urls", h.GetUserURLs)

	r.DELETE("/api/user/urls", h.DeleteUserURLs)

	return r
}

func printBuildInfo() {
	version := buildVersion
	if version == "" {
		version = "N/A"
	}

	date := buildDate
	if date == "" {
		date = "N/A"
	}

	commit := buildCommit
	if commit == "" {
		commit = "N/A"
	}

	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)
}
