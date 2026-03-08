// Package processor provides article processing orchestration.
package processor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
)

// ArticleServiceInterface defines the service methods needed for article processing.
type ArticleServiceInterface interface {
	Fetch(ctx context.Context, url string) ([]byte, error)
	Extract(ctx context.Context, htmlBytes []byte) (*model.Article, error)
	UpdateArticle(ctx context.Context, article *model.Article) error
	GetArticle(ctx context.Context, accountID, articleID string) (*model.Article, error)
}

// Processor defines the interface for starting article processing.
type Processor interface {
	StartProcessing(ctx context.Context, event *content.ProcessArticleEvent)
}

// LocalProcessor runs article processing in a goroutine.
type LocalProcessor struct {
	service ArticleServiceInterface
}

// NewLocalProcessor creates a new LocalProcessor.
func NewLocalProcessor(svc ArticleServiceInterface) *LocalProcessor {
	return &LocalProcessor{service: svc}
}

// StartProcessing starts article processing in a goroutine.
func (p *LocalProcessor) StartProcessing(ctx context.Context, event *content.ProcessArticleEvent) {
	go ProcessArticle(ctx, p.service, event)
}

// ProcessArticle processes an article: fetches, extracts, and stores it.
func ProcessArticle(
	ctx context.Context,
	svc ArticleServiceInterface,
	event *content.ProcessArticleEvent,
) {
	var requestError error
	processCtx := context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})
	processCtx = context.WithValue(processCtx, logging.RequestErrorKey, &requestError)
	processCtx, cancel := context.WithTimeout(processCtx, consts.ArticleProcessingTimeout)
	defer cancel()

	htmlBytes, err := svc.Fetch(processCtx, event.URL)
	if err != nil {
		markArticleError(processCtx, svc, event.AccountID, event.ArticleID, "fetch", err)
		logArticleResult(processCtx, event.InheritedAttrs, "failed")
		return
	}

	extractedArticle, err := svc.Extract(processCtx, htmlBytes)
	if err != nil {
		markArticleError(processCtx, svc, event.AccountID, event.ArticleID, "extract", err)
		logArticleResult(processCtx, event.InheritedAttrs, "failed")
		return
	}

	if extractedArticle == nil {
		markArticleError(processCtx, svc, event.AccountID, event.ArticleID, "extract", errors.New("extracted article is nil"))
		logArticleResult(processCtx, event.InheritedAttrs, "failed")
		return
	}

	if extractedArticle.URL != event.ArticleID {
		logging.AddLogAttr(processCtx, slog.String("url_mismatch",
			"want "+event.URL+", got "+extractedArticle.URL))
	}

	extractedArticle.Account = event.AccountID
	extractedArticle.ID = event.ArticleID
	extractedArticle.CreatedAt = time.Now().UTC()
	extractedArticle.URL = event.URL

	if updateErr := svc.UpdateArticle(processCtx, extractedArticle); updateErr != nil {
		markArticleError(processCtx, svc, event.AccountID, event.ArticleID, "update", updateErr)
		logArticleResult(processCtx, event.InheritedAttrs, "failed")
		return
	}

	logArticleResult(processCtx, event.InheritedAttrs, "success")
}

func logArticleResult(ctx context.Context, inheritedAttrs []map[string]any, status string) {
	attrs := make([]slog.Attr, len(inheritedAttrs))
	for i, attrMap := range inheritedAttrs {
		for k, v := range attrMap {
			attrs[i] = slog.Any(k, v)
		}
	}
	logging.LogArticleProcessing(ctx, "article processing completed", attrs, slog.String("status", status))
}

func markArticleError(ctx context.Context, svc ArticleServiceInterface, accountID, articleID, stage string, err error) {
	logging.AddRequestError(ctx, fmt.Errorf("article %s: %s error: %w", articleID, stage, err))

	article, getErr := svc.GetArticle(ctx, accountID, articleID)
	if getErr != nil {
		logging.AddRequestError(ctx, fmt.Errorf("failed to get article %s for error update: %w", articleID, getErr))
		return
	}

	article.Error = err.Error()
	if updateErr := svc.UpdateArticle(ctx, article); updateErr != nil {
		logging.AddRequestError(ctx, fmt.Errorf("failed to update article %s error state: %w", articleID, updateErr))
	}
}
