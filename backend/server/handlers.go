// Package server provides HTTP handlers and middleware for the savetoink application.
package server

import (
	"net/http"

	"github.com/shaftoe/savetoink/backend/config"
	"github.com/shaftoe/savetoink/backend/service"
)

func newHandlers(
	cfg *config.Config,
	svc service.Interface,
	client *http.Client,
) *handlers {
	return &handlers{
		cfg:     cfg,
		service: svc,
		client:  client,
	}
}
