// Package processor provides article processing orchestration.
package processor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/shaftoe/savetoink/backend/lib/service/servicetypes"
	"github.com/shaftoe/savetoink/backend/lib/validation"
	"golang.org/x/net/html"
)

// Service defines the interface for service operations needed by the processor.
type Service interface {
	Fetch(ctx context.Context, u *url.URL) (*content.FetchedContent, error)
	ParseHTML(ctx context.Context, fetched *content.FetchedContent) (*html.Node, error)
	Clean(ctx context.Context, doc *html.Node, u *url.URL) (*model.Article, error)
	UpdateArticle(ctx context.Context, article *model.Article) error
	GetArticle(ctx context.Context, accountID, articleID string) (*model.Article, error)
	GetUserDeviceEmail(ctx context.Context, accountID string) (email string, autoSend bool, err error)
	SendArticleByID(ctx context.Context, accountID, articleID string) (*servicetypes.SendArticleResult, error)
}

// Processor defines the interface for starting article processing.
type Processor interface {
	StartProcessing(ctx context.Context, event *content.ProcessArticleEvent)
}

// LocalProcessor runs article processing in a goroutine.
type LocalProcessor struct {
	service Service
}

// NewLocalProcessor creates a new LocalProcessor.
func NewLocalProcessor(svc Service) *LocalProcessor {
	return &LocalProcessor{service: svc}
}

// StartProcessing starts article processing in a goroutine.
func (p *LocalProcessor) StartProcessing(ctx context.Context, event *content.ProcessArticleEvent) {
	go ProcessArticle(ctx, p.service, event)
}

// ProcessArticle processes an article: fetches, extracts, stores, and optionally sends it.
//
//nolint:funlen // function has many statements due to sequential processing steps
func ProcessArticle(
	ctx context.Context,
	svc Service,
	event *content.ProcessArticleEvent,
) {
	var requestError error
	processCtx := context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})
	processCtx = context.WithValue(processCtx, logging.RequestErrorKey, &requestError)
	processCtx, cancel := context.WithTimeout(processCtx, consts.ArticleProcessingTimeout)
	defer cancel()

	logging.AddLogAttr(processCtx, slog.String("article_id", event.ArticleID))
	logging.AddLogAttr(processCtx, slog.Bool("send_on_complete", event.SendOnComplete))

	u, err := validation.ValidateURL(event.URL)
	if err != nil {
		markArticleError(processCtx, svc, event.AccountID, event.ArticleID, "parse_url", err)
		logArticleResult(processCtx, event.InheritedAttrs, "failed")
		return
	}

	fetched, err := svc.Fetch(processCtx, u)
	if err != nil {
		markArticleError(processCtx, svc, event.AccountID, event.ArticleID, "fetch", err)
		logArticleResult(processCtx, event.InheritedAttrs, "failed")
		return
	}

	logging.AddLogAttr(processCtx, slog.String("fetcher_type", fetched.Type.String()))

	doc, err := svc.ParseHTML(processCtx, fetched)
	if err != nil {
		markArticleError(processCtx, svc, event.AccountID, event.ArticleID, "parse", err)
		logArticleResult(processCtx, event.InheritedAttrs, "failed")
		return
	}

	extractedArticle, err := svc.Clean(processCtx, doc, u)
	if err != nil {
		markArticleError(processCtx, svc, event.AccountID, event.ArticleID, "clean", err)
		logArticleResult(processCtx, event.InheritedAttrs, "failed")
		return
	}

	if extractedArticle == nil {
		markArticleError(processCtx, svc, event.AccountID, event.ArticleID, "clean", errors.New("cleaned article is nil"))
		logArticleResult(processCtx, event.InheritedAttrs, "failed")
		return
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

	if event.SendOnComplete {
		if sendErr := sendArticle(processCtx, svc, event.AccountID, event.ArticleID); sendErr != nil {
			logging.AddRequestError(processCtx, sendErr)
			logArticleResult(processCtx, event.InheritedAttrs, "failed")
			return
		}
	}

	logArticleResult(processCtx, event.InheritedAttrs, "success")
}

func sendArticle(ctx context.Context, svc Service, accountID, articleID string) error {
	deviceEmail, _, err := svc.GetUserDeviceEmail(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get device email: %w", err)
	}
	if deviceEmail == "" {
		return errors.New("device email not configured")
	}

	result, err := svc.SendArticleByID(ctx, accountID, articleID)
	if err != nil {
		return fmt.Errorf("failed to send article: %w", err)
	}

	logging.AddLogAttr(ctx, slog.String("destination_email", result.DeviceEmail))
	if result.EmailResp != nil {
		logging.AddLogAttr(ctx, slog.String("message_id", result.EmailResp.MessageID))
	}

	return nil
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

func markArticleError(ctx context.Context, svc Service, accountID, articleID, stage string, err error) {
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
