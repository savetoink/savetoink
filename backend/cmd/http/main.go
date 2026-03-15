// HTTP server is the entry point for running the application as a standalone HTTP server.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/joho/godotenv"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/scheduler"
	"github.com/shaftoe/savetoink/backend/lib/server"
)

func main() {
	_ = godotenv.Load()
	cfg, err := config.Load(consts.ModeServer, func(ctx context.Context) (aws.Config, error) {
		return awsconfig.LoadDefaultConfig(ctx)
	})
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logging.SetupLogging(cfg)
	slog.Info("initializing Save to Ink API server")

	bgSched := scheduler.NewBackgroundScheduler(cfg)

	if bgSched != nil {
		if startErr := bgSched.Start(context.Background()); startErr != nil {
			slog.Error("failed to start background scheduler", "error", startErr)
			os.Exit(1)
		}
	} else {
		switch len(cfg.Tasks) {
		case 0:
			slog.Debug("no background scheduler configured")
		default:
			slog.Warn("no background scheduler configured, tasks will not run",
				slog.Any("tasks", cfg.Tasks))
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := startServer(ctx, cfg)

	<-ctx.Done()
	slog.Info("shutting down...")
	if bgSched != nil {
		bgSched.Stop()
	}

	if shutdownErr := srv.Shutdown(context.Background()); shutdownErr != nil {
		slog.Error("failed to shutdown server", slog.String("error", shutdownErr.Error()))
	} else {
		slog.Info("server shutdown successfully")
	}
}

func startServer(ctx context.Context, cfg *config.Config) *http.Server {
	var (
		port   = "8080" // TODO move to config
		router = server.NewRouter(cfg)
	)

	slog.InfoContext(ctx, "starting HTTP server", "port", port)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  consts.ReadTimeout,
		WriteTimeout: consts.WriteTimeout,
		IdleTimeout:  consts.IdleTimeout,
	}
	go func() {
		if listenErr := srv.ListenAndServe(); listenErr != nil && listenErr != http.ErrServerClosed {
			slog.ErrorContext(ctx, "failed to start server", "error", listenErr)
			os.Exit(1)
		}
	}()

	return srv
}
