package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/internal/paseto"
	"github.com/shaftoe/savetoink/backend/lib/internal/server/auth"
	"github.com/shaftoe/savetoink/backend/lib/internal/server/handlers"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/service"
)

// NewRouter creates and configures a new chi router with all middleware and routes.
func NewRouter(cfg *config.Config) *chi.Mux {
	return newRouterWithClient(cfg, &http.Client{
		Timeout: consts.Auth0ClientTimeout,
	})
}

func newRouterWithClient(cfg *config.Config, client *http.Client) *chi.Mux {
	r := chi.NewRouter()
	srv := service.NewFromConfig(cfg)
	proc := newProcessor(cfg, srv)

	var keyStore *paseto.KeyStore
	if cfg.AuthBackend == consts.AuthBackendAuth0 {
		keyStore = newPasetoKeyStore(cfg)
	}

	h := handlers.New(
		cfg,
		srv,
		client,
		proc,
		keyStore,
	)

	r.Use(middleware.Recoverer)
	r.Use(auth.NewAccountIDMiddleware(cfg.AuthBackend, cfg.APIKeySecret, keyStore))
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

	r.Get("/robots.txt", handlers.RobotsTXTHandler)
	r.Get("/v1/openapi.yaml", handlers.OpenAPIHandler)

	setupRoutes(r, h, cfg, srv)

	return r
}

func setupRoutes(r *chi.Mux, h *handlers.Handlers, cfg *config.Config, _ service.Interface) {
	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", h.HandleHealth)

		r.Route("/articles", func(r chi.Router) {
			r.Use(auth.EnsureAutheticatedMiddleware)
			r.With(processorInfoMiddleware(cfg)).Post("/", h.HandleCreateArticle)
			r.Get("/", h.HandleGetArticles)
			r.Get("/{id}", h.HandleGetArticle)
			r.Delete("/{id}", h.HandleDeleteArticle)
			r.Put("/{id}/favorite", h.HandleToggleFavorite)
			r.Post("/{id}/send", h.HandleSendArticle)
			r.Post("/{id}/tags", h.HandleAddTags)
			r.Put("/{id}/tags", h.HandleSetTags)
			r.Get("/{id}/tags", h.HandleGetTags)
			r.Delete("/{id}/tags", h.HandleRemoveTags)
		})

		r.Route("/user", func(r chi.Router) {
			r.Use(auth.EnsureAutheticatedMiddleware)
			r.Route("/profile", func(r chi.Router) {
				r.Get("/", h.HandleGetUserProfile)
			})
		})

		r.Route("/devices", func(r chi.Router) {
			r.Use(auth.EnsureAutheticatedMiddleware)
			r.Put("/", h.HandleSetDevice)
			r.Delete("/", h.HandleDeleteDevice)
		})

		if cfg.AuthBackend == consts.AuthBackendAuth0 {
			r.Post("/auth/token", h.HandleAuthTokenExchange)
			r.Post("/auth/refresh", h.HandleAuthRefresh)
		}

		if cfg.EmailProvider == consts.EmailBackendMailjet {
			r.Post("/webhooks/mailjet", h.HandleMailjetWebhook)
		}

		r.Route("/sends", func(r chi.Router) {
			r.Use(auth.EnsureAutheticatedMiddleware)
			r.Get("/", h.HandleGetSends)
		})

		r.Route("/tags", func(r chi.Router) {
			r.Use(auth.EnsureAutheticatedMiddleware)
			r.Get("/", h.HandleGetAllTags)
		})
	})
}

func newPasetoKeyStore(cfg *config.Config) *paseto.KeyStore {
	keyStore, err := paseto.NewKeyStore(paseto.KeyStoreConfig{
		SymmetricKey: cfg.PASETOSymmetricKey,
		KeyVersion:   cfg.PASETOKeyVersion,
	})
	if err != nil {
		panic("failed to create paseto key store: " + err.Error())
	}
	return keyStore
}
