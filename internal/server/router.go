package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/shaftoe/savetoink/internal/config"
	"github.com/shaftoe/savetoink/internal/consts"
	"github.com/shaftoe/savetoink/internal/model"
	"github.com/shaftoe/savetoink/internal/server/auth"
	"github.com/shaftoe/savetoink/internal/service"
)

// NewRouter creates and configures a new chi router with all middleware and routes.
func NewRouter(cfg *config.Config) *chi.Mux {
	return newRouterWithClient(cfg, &http.Client{
		Timeout: consts.Auth0ClientTimeout,
	})
}

func newRouterWithClient(cfg *config.Config, client *http.Client) *chi.Mux {
	setupLogging(cfg)

	r := chi.NewRouter()
	srv := service.New(cfg)

	handlers := newHandlers(
		cfg,
		srv,
		client,
	)

	r.Use(middleware.Recoverer)
	r.Use(auth.NewAccountIDMiddleware(cfg))
	r.Use(requestIDMiddleware)
	r.Use(loggingMiddleware)
	r.Use(corsMiddleware)
	r.Use(jsonContentTypeMiddleware)

	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "not_found"})
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "method_not_allowed"})
	})

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", handlers.handleHealth)

		r.Route("/articles", func(r chi.Router) {
			r.Use(auth.EnsureAutheticatedMiddleware)
			r.Post("/", handlers.handleCreateArticle)
			r.Get("/", handlers.handleGetArticles)
			r.Delete("/", handlers.handleDeleteAllArticles)
			r.Get("/{id}", handlers.handleGetArticle)
			r.Delete("/{id}", handlers.handleDeleteArticle)
			r.Put("/{id}/favorite", handlers.handleToggleFavorite)
		})

		r.Route("/user", func(r chi.Router) {
			r.Use(auth.EnsureAutheticatedMiddleware)
			r.Route("/profile", func(r chi.Router) {
				r.Get("/", handlers.handleGetUserProfile)
				r.Put("/", handlers.handleSetUserProfile)
				r.Delete("/", handlers.handleDeleteProfile)
			})
		})

		if cfg.AuthBackend == consts.AuthBackendAuth0 {
			r.Post("/auth/token", handlers.handleAuthTokenExchange)
		}
	})

	return r
}

func setupLogging(cfg *config.Config) {
	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))
}

const logRecordKey = contextKey("log_record")

func addLogAttr(ctx context.Context, attr slog.Attr) {
	if record, ok := ctx.Value(logRecordKey).(*logRecord); ok {
		record.AddAttrs(attr)
	}
}
