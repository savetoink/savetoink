// Package server provides HTTP handlers and middleware for the savetoink application.
package server

import (
	_ "embed"
	"net/http"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/processor"
	"github.com/shaftoe/savetoink/backend/lib/service"
)

//go:embed openapi.yaml
var openapi string

func newHandlers(
	cfg *config.Config,
	svc service.Interface,
	client *http.Client,
	proc processor.Processor,
) *handlers {
	return &handlers{
		cfg:       cfg,
		service:   svc,
		client:    client,
		processor: proc,
	}
}

func robotsTXTHandler(w http.ResponseWriter, _ *http.Request) {
	const robotsTxt = `User-agent: *
Allow: /v1/openapi.yaml
Disallow: /`
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(robotsTxt))
}

func openAPIHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/yaml")
	_, _ = w.Write([]byte(openapi))
}
