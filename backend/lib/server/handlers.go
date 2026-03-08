// Package server provides HTTP handlers and middleware for the savetoink application.
package server

import (
	"net/http"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/processor"
	"github.com/shaftoe/savetoink/backend/lib/service"
)

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
