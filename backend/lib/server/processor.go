package server

import (
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/processor"
	"github.com/shaftoe/savetoink/backend/lib/processor/lambda"
	"github.com/shaftoe/savetoink/backend/lib/service"
)

func newProcessor(cfg *config.Config, svc service.Interface) processor.Processor {
	if cfg.ProcessArticleLambda != "" {
		return lambda.NewProcessor(cfg.ProcessArticleLambda, cfg.AWSConfig)
	}
	return processor.NewLocalProcessor(svc)
}
