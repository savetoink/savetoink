// Package processor provides article processing orchestration.
package processor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/email"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
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
	GenerateEPUB(article *model.Article) (io.ReadCloser, error)
	GetUserDeviceEmailAndAutoSend(ctx context.Context, accountID string) (email string, autoSend bool, err error)
	SendArticle(
		ctx context.Context,
		destEmail string,
		epubData io.ReadCloser,
		title string,
	) (*email.SendEmailResponse, error)
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
func ProcessArticle(
	ctx context.Context,
	svc Service,
	event *content.ProcessArticleEvent,
) {
	processCtx, cancel := setupProcessingContext(ctx, event)
	defer cancel()

	u, validateErr := validateArticleURL(event.URL)
	if validateErr != nil {
		handleProcessingError(processCtx, svc, event, "parse_url", validateErr)
		return
	}

	fetched, fetchErr := fetchArticleContent(processCtx, u, svc)
	if fetchErr != nil {
		handleProcessingError(processCtx, svc, event, "fetch", fetchErr)
		return
	}

	doc, parseErr := svc.ParseHTML(processCtx, fetched)
	if parseErr != nil {
		handleProcessingError(processCtx, svc, event, "parse", parseErr)
		return
	}

	article, cleanErr := cleanArticle(processCtx, doc, u, svc)
	if cleanErr != nil {
		handleProcessingError(processCtx, svc, event, "clean", cleanErr)
		return
	}

	if article == nil {
		handleProcessingError(processCtx, svc, event, "clean", errors.New("cleaned article is nil"))
		return
	}

	updatedArticle, storeErr := prepareAndStoreArticle(processCtx, article, event, svc)
	if storeErr != nil {
		handleProcessingError(processCtx, svc, event, "update", storeErr)
		return
	}

	if event.SendOnComplete {
		sendErr := sendArticle(processCtx, svc, updatedArticle)
		if sendErr != nil {
			logging.AddRequestError(processCtx, sendErr)
			logArticleResult(processCtx, event.InheritedAttrs, "failed")
			return
		}
	}

	logArticleResult(processCtx, event.InheritedAttrs, "success")
}

func setupProcessingContext(
	ctx context.Context,
	event *content.ProcessArticleEvent,
) (context.Context, context.CancelFunc) {
	var requestError error
	processCtx := context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})
	processCtx = context.WithValue(processCtx, logging.RequestErrorKey, &requestError)
	processCtx, cancel := context.WithTimeout(processCtx, consts.ArticleProcessingTimeout)
	logging.AddLogAttr(processCtx, slog.String("article_id", event.ArticleID))
	logging.AddLogAttr(processCtx, slog.Bool("send_on_complete", event.SendOnComplete))
	return processCtx, cancel
}

func validateArticleURL(urlStr string) (*url.URL, error) {
	u, err := validation.ValidateURL(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to validate URL: %w", err)
	}
	return u, nil
}

func fetchArticleContent(
	ctx context.Context,
	u *url.URL,
	svc Service,
) (*content.FetchedContent, error) {
	fetched, err := svc.Fetch(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch content: %w", err)
	}
	logging.AddLogAttr(ctx, slog.String("fetcher_type", fetched.Type.String()))
	return fetched, nil
}

func cleanArticle(
	ctx context.Context,
	doc *html.Node,
	u *url.URL,
	svc Service,
) (*model.Article, error) {
	article, err := svc.Clean(ctx, doc, u)
	if err != nil {
		return nil, fmt.Errorf("failed to clean article: %w", err)
	}
	return article, nil
}

func prepareAndStoreArticle(
	ctx context.Context,
	article *model.Article,
	event *content.ProcessArticleEvent,
	svc Service,
) (*model.Article, error) {
	article.Account = event.AccountID
	article.ID = event.ArticleID
	article.CreatedAt = time.Now().UTC()
	article.URL = event.URL

	if err := svc.UpdateArticle(ctx, article); err != nil {
		return nil, fmt.Errorf("failed to update article: %w", err)
	}
	return article, nil
}

func handleProcessingError(
	ctx context.Context,
	svc Service,
	event *content.ProcessArticleEvent,
	stage string,
	err error,
) {
	markArticleError(ctx, svc, event.AccountID, event.ArticleID, stage, err)
	logArticleResult(ctx, event.InheritedAttrs, "failed")
}

func sendArticle(ctx context.Context, svc Service, article *model.Article) error {
	deviceEmail, _, err := svc.GetUserDeviceEmailAndAutoSend(ctx, article.Account)
	if err != nil {
		return fmt.Errorf("failed to get device email: %w", err)
	}
	if deviceEmail == "" {
		return errors.New("device email not configured")
	}

	epub, err := svc.GenerateEPUB(article)
	if err != nil {
		return fmt.Errorf("failed to generate epub: %w", err)
	}
	defer func() {
		if closeErr := epub.Close(); closeErr != nil {
			logging.AddRequestError(ctx, fmt.Errorf("failed to close epub: %w", closeErr))
		}
	}()

	result, err := svc.SendArticle(ctx, deviceEmail, epub, article.Title)
	if err != nil {
		return fmt.Errorf("failed to send article: %w", err)
	}

	logging.AddLogAttr(ctx, slog.String("destination_email", deviceEmail))
	if result.MessageID != "" {
		logging.AddLogAttr(ctx, slog.String("message_id", result.MessageID))
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
