// HTTP server is the entry point for running the application as a standalone HTTP server.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/shaftoe/savetoink/internal/config"
	"github.com/shaftoe/savetoink/internal/consts"
	"github.com/shaftoe/savetoink/internal/server"
)

func main() {
	cfg, err := config.Load(consts.ModeServer)
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	var (
		version = consts.Version()
		port    = "8080"
		router  = server.NewRouter(cfg)
	)

	slog.Info("starting Save to Ink development HTTP server", "port", port, "version", *version)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  consts.ReadTimeout,
		WriteTimeout: consts.WriteTimeout,
		IdleTimeout:  consts.IdleTimeout,
	}
	if srvErr := srv.ListenAndServe(); srvErr != nil {
		slog.Error("failed to start server", "error", srvErr)
		os.Exit(1)
	}
}
