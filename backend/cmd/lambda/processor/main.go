// Lambda processor handles async article processing.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/processor"
	"github.com/shaftoe/savetoink/backend/lib/service"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
)

func main() {
	cfg, err := config.Load(consts.ModeServer, func(ctx context.Context) (aws.Config, error) {
		return awsconfig.LoadDefaultConfig(ctx)
	})
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logging.SetupLogging(cfg)
	svc := service.NewFromConfig(cfg)
	lambda.Start(func(ctx context.Context, event *content.ProcessArticleEvent) error {
		if event == nil {
			event = &content.ProcessArticleEvent{}
		}
		return handleRequest(ctx, event, svc)
	})
}

func handleRequest(ctx context.Context, event *content.ProcessArticleEvent, svc service.Interface) error {
	if event.RequestID == "" {
		slog.Error("invalid request: missing request_id")
		return nil
	}

	if event.URL == "" {
		slog.Error("invalid request: missing url", "request_id", event.RequestID)
		return nil
	}

	if event.ArticleID == "" {
		slog.Error("invalid request: missing article_id", "request_id", event.RequestID)
		return nil
	}

	if event.AccountID == "" {
		slog.Error("invalid request: missing account_id", "request_id", event.RequestID)
		return nil
	}

	processCtx := context.WithValue(ctx, logging.RequestIDKey, event.RequestID)
	processCtx = context.WithValue(processCtx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})

	processor.ProcessArticle(processCtx, svc, event)

	return nil
}
