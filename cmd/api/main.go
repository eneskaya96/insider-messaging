package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/eneskaya/insider-messaging/docs"
	"github.com/eneskaya/insider-messaging/internal/application/service"
	"github.com/eneskaya/insider-messaging/internal/infrastructure/cache"
	infrahttp "github.com/eneskaya/insider-messaging/internal/infrastructure/http"
	"github.com/eneskaya/insider-messaging/internal/infrastructure/persistence"
	"github.com/eneskaya/insider-messaging/internal/infrastructure/scheduler"
	"github.com/eneskaya/insider-messaging/internal/presentation/handler"
	"github.com/eneskaya/insider-messaging/internal/presentation/router"
	"github.com/eneskaya/insider-messaging/pkg/config"
	"github.com/eneskaya/insider-messaging/pkg/logger"
	"go.uber.org/zap"
)

// @title Insider Messaging API
// @version 1.0
// @description Automatic message sending system with scheduler
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@insider.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := logger.Init(cfg.App.LogLevel); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	logger.Get().Info("starting application",
		zap.String("env", cfg.App.Env),
		zap.String("port", cfg.App.Port),
	)

	db, err := persistence.NewPostgresGormDB(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	redisCache, err := cache.NewRedisCache(&cfg.Redis)
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	defer redisCache.Close()

	messageCache := cache.NewMessageCache(redisCache)

	webhookClient := infrahttp.NewWebhookClient(&cfg.Webhook)

	messageRepo := persistence.NewMessageRepositoryGorm(db.DB(), cfg.Message.CharLimit)

	messageService := service.NewMessageService(
		messageRepo,
		webhookClient,
		messageCache,
		cfg.Message.CharLimit,
		cfg.Message.MaxRetries,
	)

	msgScheduler := scheduler.NewScheduler(
		messageService,
		cfg.Message.BatchSize,
		cfg.Message.IntervalSeconds,
		cfg.Message.WorkerCount,
	)

	messageHandler := handler.NewMessageHandler(messageService)
	schedulerHandler := handler.NewSchedulerHandler(msgScheduler)
	healthHandler := handler.NewHealthHandler(db, redisCache)

	r := router.NewRouter(messageHandler, schedulerHandler, healthHandler)
	engine := r.Setup()

	srv := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: engine,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := msgScheduler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	go func() {
		logger.Get().Info("starting HTTP server", zap.String("port", cfg.App.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Get().Fatal("failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Get().Info("shutting down application...")

	if err := msgScheduler.Stop(); err != nil {
		logger.Get().Error("error stopping scheduler", zap.Error(err))
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.App.GracefulShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Get().Error("server forced to shutdown", zap.Error(err))
		return err
	}

	logger.Get().Info("application stopped gracefully")
	return nil
}
