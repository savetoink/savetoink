package lambda

import (
	"context"
	"log/slog"

	lambdacontext "github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/processor"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
)

// HandleEvent is the entry point for processing a single article event via AWS Lambda.
func HandleEvent(ctx context.Context, event *content.ProcessArticleEvent, svc processor.Service) error {
	if event == nil {
		event = &content.ProcessArticleEvent{}
	}

	lc, _ := lambdacontext.FromContext(ctx)
	lambdaRequestID := lc.AwsRequestID
	logger := slog.Default().With(slog.String("request_id", lambdaRequestID)).With("version", *consts.Version())

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
		map[string]any{"account_id": event.AccountID},
		map[string]any{"article_id": event.ArticleID},
		map[string]any{"orig_request_id": event.RequestID},
		map[string]any{"request_id": lambdaRequestID},
		map[string]any{"url": event.URL},
		map[string]any{"version": *consts.Version()},
	)
	event.RequestID = lambdaRequestID

	processCtx := logging.WithRequestID(ctx, event.RequestID)
	processCtx = logging.WithLogRecord(processCtx)

	processor.ProcessArticle(processCtx, svc, event)

	return nil
}
