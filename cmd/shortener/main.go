package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/Guldana11/shortener/api/shortener"
	"github.com/Guldana11/shortener/internal/audit"
	"github.com/Guldana11/shortener/internal/config"
	"github.com/Guldana11/shortener/internal/config/db"
	"github.com/Guldana11/shortener/internal/grpcserver"
	"github.com/Guldana11/shortener/internal/handler"
	"github.com/Guldana11/shortener/internal/middleware"
	"github.com/Guldana11/shortener/internal/repository"
	"github.com/Guldana11/shortener/internal/service"
	"github.com/Guldana11/shortener/internal/worker"
	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	printBuildInfo()

	go func() {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			zap.L().Fatal("pprof server failed", zap.Error(err))
		}
	}()

	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("failed to init config: %v", err)
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	var (
		repo      repository.Repository
		poolClose func()
	)

	if cfg.DatabaseDSN != "" {
		pool, err := db.NewPostgresPool(context.Background(), cfg.DatabaseDSN, "file://migrations")
		if err != nil {
			logger.Fatal("failed to connect to database", zap.Error(err))
		}
		poolClose = pool.Close

		repo = repository.NewPostgresRepository(pool)
		logger.Info("Using PostgreSQL storage")
	} else if cfg.FileStoragePath != "" {
		repo = repository.NewURLRepository(cfg.FileStoragePath)
		logger.Info("Using file storage", zap.String("file", cfg.FileStoragePath))
	} else {
		repo = repository.NewURLRepository("")
		logger.Info("Using in-memory storage")
	}

	// гарантируем закрытие пула БД при остановке
	defer func() {
		if poolClose != nil {
			poolClose()
		}
	}()

	publisher := audit.NewPublisher()

	if cfg.AuditFile != "" {
		publisher.Subscribe(audit.NewFileObserver(cfg.AuditFile))
	}
	if cfg.AuditURL != "" {
		publisher.Subscribe(audit.NewHTTPObserver(cfg.AuditURL))
	}

	deleteWorker := worker.NewDeleteWorker(repo, 1000)

	svc := &service.URLService{
		Repo:      repo,
		BaseURL:   cfg.BaseURL,
		Publisher: publisher,
		Logger:    logger,
	}

	h := handler.NewURLHandler(cfg.BaseURL, repo, deleteWorker, svc, cfg.TrustedSubnet)
	r := setupRouter(h, logger)
	logger.Info("Server is starting...", zap.String("address", cfg.Address))
	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	logger.Info("Server is starting...",
		zap.String("address", cfg.Address),
		zap.Bool("https", cfg.EnableHTTPS),
	)

	// запускаем HTTP-сервер в отдельной горутине
	errCh := make(chan error, 1)
	go func() {
		if cfg.EnableHTTPS {
			errCh <- srv.ListenAndServeTLS("cert.pem", "key.pem")
			return
		}
		errCh <- srv.ListenAndServe()
	}()

	// запускаем gRPC-сервер
	grpcOpts := []grpc.ServerOption{grpc.UnaryInterceptor(grpcserver.AuthInterceptor())}
	if cfg.EnableHTTPS {
		creds, err := credentials.NewServerTLSFromFile("cert.pem", "key.pem")
		if err != nil {
			logger.Fatal("failed to load TLS credentials for gRPC", zap.Error(err))
		}
		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}
	grpcSrv := grpc.NewServer(grpcOpts...)
	pb.RegisterShortenerServiceServer(grpcSrv, &grpcserver.ShortenerServer{Svc: svc})

	grpcLis, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		logger.Fatal("failed to listen for gRPC", zap.Error(err))
	}
	go func() {
		logger.Info("gRPC server is starting...", zap.String("address", cfg.GRPCAddress))
		if err := grpcSrv.Serve(grpcLis); err != nil {
			logger.Error("gRPC server failed", zap.Error(err))
		}
	}()

	// ловим сигналы завершения
	sigCtx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT,
	)
	defer stop()

	// ждём либо сигнал, либо падение сервера
	select {
	case <-sigCtx.Done():
		logger.Info("Shutdown signal received")
	case err := <-errCh:
		// если сервер сам завершился
		if err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
		// http.ErrServerClosed = норм при shutdown
		return
	}

	// graceful shutdown: дождаться активных запросов
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown failed", zap.Error(err))
	} else {
		logger.Info("HTTP server stopped gracefully")
	}

	grpcSrv.GracefulStop()
	logger.Info("gRPC server stopped gracefully")

	deleteWorker.Stop()

	// сохранить несохранённые данные (актуально для file/in-memory репо)
	if s, ok := repo.(interface{ Save() error }); ok {
		if err := s.Save(); err != nil {
			logger.Error("Failed to save repository", zap.Error(err))
		} else {
			logger.Info("Repository saved")
		}
	}

	logger.Info("Shutdown complete")
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
	r.GET("/api/internal/stats", h.StatsHandler)

	return r
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
