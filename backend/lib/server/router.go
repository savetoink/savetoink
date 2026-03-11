package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/server/auth"
	"github.com/shaftoe/savetoink/backend/lib/service"
)

// NewRouter creates and configures a new chi router with all middleware and routes.
func NewRouter(cfg *config.Config) *chi.Mux {
	return newRouterWithClient(cfg, &http.Client{
		Timeout: consts.Auth0ClientTimeout,
	})
}

func newRouterWithClient(cfg *config.Config, client *http.Client) *chi.Mux {
	logging.SetupLogging(cfg)

	r := chi.NewRouter()
	srv := service.NewFromConfig(cfg)
	proc := newProcessor(cfg, srv)

	handlers := newHandlers(
		cfg,
		srv,
		client,
		proc,
	)

	r.Use(middleware.Recoverer)
	r.Use(auth.NewAccountIDMiddleware(cfg))
	r.Use(requestIDMiddleware)
	r.Use(logging.Middleware)
	r.Use(newCorsMiddleware(cfg))
	r.Use(jsonContentTypeMiddleware)

	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "not_found"})
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "method_not_allowed"})
	})

	r.Get("/robots.txt", robotsTXTHandler)
	r.Get("/v1/openapi.yaml", openAPIHandler)

	setupRoutes(r, handlers, cfg, srv)

	return r
}

func setupRoutes(r *chi.Mux, handlers *handlers, cfg *config.Config, _ service.Interface) {
	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", handlers.handleHealth)

		r.Route("/articles", func(r chi.Router) {
			r.Use(auth.EnsureAutheticatedMiddleware)
			r.With(processorInfoMiddleware(cfg)).Post("/", handlers.handleCreateArticle)
			r.Get("/", handlers.handleGetArticles)
			r.Delete("/", handlers.handleDeleteAllArticles)
			r.Get("/{id}", handlers.handleGetArticle)
			r.Delete("/{id}", handlers.handleDeleteArticle)
			r.Put("/{id}/favorite", handlers.handleToggleFavorite)
			r.Post("/{id}/send", handlers.handleSendArticle)
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
