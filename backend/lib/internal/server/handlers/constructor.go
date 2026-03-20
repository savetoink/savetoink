package handlers

import (
	"net/http"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/processor"
	"github.com/shaftoe/savetoink/backend/lib/service"
)

// New creates a new Handlers instance.
func New(
	cfg *config.Config,
	svc service.Interface,
	client *http.Client,
	proc processor.Processor,
) *Handlers {
	return &Handlers{
		cfg:       cfg,
		service:   svc,
		client:    client,
		processor: proc,
	}
}
