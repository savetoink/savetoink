// Lambda processor handles async article processing.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	lambdacontext "github.com/aws/aws-lambda-go/lambdacontext"
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
	lc, _ := lambdacontext.FromContext(ctx)
	lambdaRequestID := lc.AwsRequestID
	logger := slog.Default().With(slog.String("request_id", lambdaRequestID))

	if event.RequestID == "" {
		logger.Error("invalid request: missing orig_request_id")
		return nil
	}

	logger = logger.With(slog.String("orig_request_id", event.RequestID))

	if event.URL == "" {
		logger.Error("invalid request: missing url")
		return nil
	}

	if event.ArticleID == "" {
		logger.Error("invalid request: missing article_id")
		return nil
	}

	if event.AccountID == "" {
		logger.Error("invalid request: missing account_id")
		return nil
	}

	event.InheritedAttrs = make([]map[string]any, 0, len(event.InheritedAttrs))
	for _, attr := range event.InheritedAttrs {
		if _, exists := attr["request_id"]; !exists {
			event.InheritedAttrs = append(event.InheritedAttrs, attr)
		}
	}
	event.InheritedAttrs = append(event.InheritedAttrs,
		map[string]any{"orig_request_id": event.RequestID},
		map[string]any{"request_id": lambdaRequestID},
	)
	event.RequestID = lambdaRequestID

	processCtx := context.WithValue(ctx, logging.RequestIDKey, event.RequestID)
	processCtx = context.WithValue(processCtx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})

	processor.ProcessArticle(processCtx, svc, event)

	return nil
}
