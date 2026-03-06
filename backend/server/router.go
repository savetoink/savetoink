package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentryslog "github.com/getsentry/sentry-go/slog"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/shaftoe/savetoink/backend/config"
	"github.com/shaftoe/savetoink/backend/consts"
	"github.com/shaftoe/savetoink/backend/model"
	"github.com/shaftoe/savetoink/backend/server/auth"
	"github.com/shaftoe/savetoink/backend/service"
)

const sentryTimeout = 5 * time.Second

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

	setupRoutes(r, handlers, cfg, srv)

	return r
}

func setupRoutes(r *chi.Mux, handlers *handlers, cfg *config.Config, srv service.Interface) {
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
			r.With(
				auth.NewEmailBackendEnabledMiddleware(cfg),
				auth.NewActiveSubscriptionMiddleware(cfg, srv),
				auth.NewBouncingEmailMiddleware(srv),
			).Post("/{id}/send", handlers.handleSendArticle)
		})

		r.Route("/user", func(r chi.Router) {
			r.Use(auth.EnsureAutheticatedMiddleware)
			r.Route("/profile", func(r chi.Router) {
				r.Get("/", handlers.handleGetUserProfile)
			})
		})

		r.Route("/devices", func(r chi.Router) {
			r.Use(auth.EnsureAutheticatedMiddleware)
			r.Put("/", handlers.handleSetDevice)
			r.Delete("/", handlers.handleDeleteDevice)
		})

		if cfg.AuthBackend == consts.AuthBackendAuth0 {
			r.Post("/auth/token", handlers.handleAuthTokenExchange)
		}

		if cfg.EmailProvider == consts.EmailBackendMailjet {
			r.Post("/webhooks/mailjet", handlers.handleMailjetWebhook)
		}
	})
}

func setupLogging(cfg *config.Config) {
	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}

	defaultHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	slog.SetDefault(slog.New(defaultHandler))

	if cfg.LoggingProvider == consts.LoggingBackendNone {
		return
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Debug:            cfg.Debug,
		Environment:      cfg.SentryEnvironment,
		SampleRate:       cfg.SentrySampleRate,
		AttachStacktrace: true,
		EnableLogs:       true,
		Transport: &sentry.HTTPSyncTransport{
			Timeout: sentryTimeout,
		},
	})

	if err != nil {
		slog.Error("failed to initialize Sentry, fall back to default logger", "error", err)

		return
	}

	logLevels := []slog.Level{slog.LevelInfo, slog.LevelWarn, slog.LevelError}
	if cfg.Debug {
		logLevels = append(logLevels, slog.LevelDebug)
	}

	sentryHandler := sentryslog.Option{
		EventLevel: []slog.Level{slog.LevelWarn, slog.LevelError, sentryslog.LevelFatal},
		LogLevel:   logLevels,
	}.NewSentryHandler(context.Background())

	multiHandler := slog.NewMultiHandler(sentryHandler, defaultHandler)
	slog.SetDefault(slog.New(multiHandler))
}

const logRecordKey = contextKey("log_record")

func addLogAttr(ctx context.Context, attr slog.Attr) {
	if record, ok := ctx.Value(logRecordKey).(*logRecord); ok {
		record.AddAttrs(attr)
	}
}

func addRequestError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	if errPtr, ok := ctx.Value(requestErrorKey).(*error); ok && errPtr != nil {
		if *errPtr != nil {
			*errPtr = errors.Join(*errPtr, err)
		} else {
			*errPtr = err
		}
	}
}

func getRequestError(ctx context.Context) error {
	if errPtr, ok := ctx.Value(requestErrorKey).(*error); ok && errPtr != nil {
		return *errPtr
	}
	return nil
}
