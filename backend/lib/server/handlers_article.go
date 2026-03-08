// Package server provides HTTP handlers and middleware for the savetoink application.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/server/auth"
	"github.com/shaftoe/savetoink/backend/lib/validation"
)

const (
	// articleProcessingTimeout is the timeout for background article processing.
	articleProcessingTimeout = 5 * time.Minute
)

func (h *handlers) handleCreateArticle(w http.ResponseWriter, r *http.Request) {
	var req articleRequest
	if err := decodeAndValidateRequest(w, r, &req); err != nil {
		return
	}

	if req.URL == "" {
		writeJSONError(w, http.StatusBadRequest, errors.New("missing URL in request body"))
		return
	}

	if err := validation.ValidateURL(req.URL); err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	logging.AddLogAttr(r.Context(), slog.String("url", req.URL))

	article, err := h.service.CreateArticle(r.Context(), req.URL, auth.GetAccountID(r.Context()))
	if err != nil {
		handleServiceError(w, r, err, "create article")
		return
	}

	logging.AddLogAttr(r.Context(), slog.String("article_id", article.ID))

	if dbErr := h.service.GetDBError(); dbErr != nil {
		logging.AddRequestError(r.Context(), fmt.Errorf("db error: %w", dbErr))
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(articleResponse{
		ID:    article.ID,
		Title: article.Title,
		URL:   article.URL,
	})

	inheritedAttrs := extractInheritedLogAttrs(r.Context())
	accountID := auth.GetAccountID(r.Context())
	url := req.URL
	articleID := article.ID

	h.startArticleProcessing(url, articleID, accountID, inheritedAttrs)
}

func (h *handlers) startArticleProcessing(url, articleID, accountID string, inheritedAttrs []slog.Attr) {
	go h.processArticleAsync(url, articleID, accountID, inheritedAttrs)
}

func extractInheritedLogAttrs(ctx context.Context) []slog.Attr {
	logRecord, ok := ctx.Value(logging.LogRecordKey).(*logging.LogRecord)
	if !ok || logRecord == nil {
		return nil
	}

	var attrs []slog.Attr
	excludeKeys := map[string]bool{
		"client_ip":  true,
		"user_agent": true,
		"path":       true,
		"method":     true,
		"url":        true,
	}
	logRecord.Attrs(func(a slog.Attr) bool {
		if !excludeKeys[a.Key] {
			attrs = append(attrs, a)
		}
		return true
	})
	return attrs
}

// processArticleAsync processes article in background goroutine.
func (h *handlers) processArticleAsync(url, articleID, accountID string, inheritedAttrs []slog.Attr) {
	var requestError error
	processCtx := context.Background()
	processCtx = context.WithValue(processCtx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})
	processCtx = context.WithValue(processCtx, logging.RequestErrorKey, &requestError)
	processCtx, cancel := context.WithTimeout(processCtx, articleProcessingTimeout)
	defer cancel()

	htmlBytes, err := h.service.Fetch(processCtx, url)
	if err != nil {
		h.markArticleError(processCtx, accountID, articleID, "fetch", err)
		h.logArticleProcessing(processCtx, inheritedAttrs, slog.String("status", "failed"))
		return
	}

	extractedArticle, err := h.service.Extract(processCtx, htmlBytes)
	if err != nil {
		h.markArticleError(processCtx, accountID, articleID, "extract", err)
		h.logArticleProcessing(processCtx, inheritedAttrs, slog.String("status", "failed"))
		return
	}

	if extractedArticle == nil {
		h.markArticleError(processCtx, accountID, articleID, "extract", errors.New("extracted article is nil"))
		h.logArticleProcessing(processCtx, inheritedAttrs, slog.String("status", "failed"))
		return
	}

	if extractedArticle.URL != articleID {
		logging.AddLogAttr(processCtx, slog.String("url_mismatch",
			fmt.Sprintf("want %s, got %s", url, extractedArticle.URL)))
	}

	extractedArticle.Account = accountID
	extractedArticle.ID = articleID
	extractedArticle.CreatedAt = time.Now().UTC()
	extractedArticle.URL = url

	if updateErr := h.service.UpdateArticle(processCtx, extractedArticle); updateErr != nil {
		h.markArticleError(processCtx, accountID, articleID, "update", updateErr)
		h.logArticleProcessing(processCtx, inheritedAttrs, slog.String("status", "failed"))
		return
	}

	h.logArticleProcessing(processCtx, inheritedAttrs, slog.String("status", "success"))
}

func (h *handlers) logArticleProcessing(ctx context.Context, inheritedAttrs []slog.Attr, extraAttr slog.Attr) {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "article processing completed", 0)
	for _, attr := range inheritedAttrs {
		record.AddAttrs(attr)
	}
	record.AddAttrs(extraAttr)

	if requestError := logging.GetRequestError(ctx); requestError != nil {
		if joinedErr, ok := requestError.(interface{ Unwrap() []error }); ok {
			for i, err := range joinedErr.Unwrap() {
				record.AddAttrs(slog.String(fmt.Sprintf("error_%d", i), err.Error()))
			}
		} else {
			record.AddAttrs(slog.String("error", requestError.Error()))
		}
		record.Level = slog.LevelError
	}

	if err := slog.Default().Handler().Handle(ctx, record); err != nil {
		slog.Error("failed to log article processing", "error", err)
	}
}

// markArticleError logs error and updates article.Error field in DB.
func (h *handlers) markArticleError(ctx context.Context, accountID, articleID, stage string, err error) {
	logging.AddRequestError(ctx, fmt.Errorf("article %s: %s error: %w", articleID, stage, err))

	article, getErr := h.service.GetArticle(ctx, accountID, articleID)
	if getErr != nil {
		logging.AddRequestError(ctx, fmt.Errorf("failed to get article %s for error update: %w", articleID, getErr))
		return
	}

	article.Error = err.Error()
	if updateErr := h.service.UpdateArticle(ctx, article); updateErr != nil {
		logging.AddRequestError(ctx, fmt.Errorf("failed to update article %s error state: %w", articleID, updateErr))
	}
}

func (h *handlers) handleGetArticles(w http.ResponseWriter, r *http.Request) {
	page := consts.DefaultPage
	pageSize := consts.DefaultPageSize
	var favoriteFilter *bool

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed >= consts.MinPage {
			page = parsed
		}
	}

	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed >= consts.MinPageSize {
			pageSize = min(parsed, consts.MaxPageSize)
		}
	}

	if f := r.URL.Query().Get("favorite"); f == "true" {
		fav := true
		favoriteFilter = &fav
	}

	accountID := auth.GetAccountID(r.Context())

	result, err := h.service.GetArticlesMetadata(r.Context(), accountID, page, pageSize, favoriteFilter)
	if err != nil {
		handleServiceError(w, r, err, "get articles metadata")
		return
	}

	logging.AddLogAttr(r.Context(), slog.Int("page", page))
	logging.AddLogAttr(r.Context(), slog.Int("page_size", pageSize))
	logging.AddLogAttr(r.Context(), slog.Int("total", result.Total))

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(listArticlesResponse{
		Articles: result.Articles,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
		HasMore:  result.HasMore,
	})
}

func (h *handlers) handleGetArticle(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountID(r.Context())
	articleID := chi.URLParam(r, "id")

	logging.AddLogAttr(r.Context(), slog.String("article_id", articleID))

	article, err := h.service.GetArticle(r.Context(), accountID, articleID)
	if err != nil {
		logging.AddRequestError(r.Context(), fmt.Errorf("db error: %w", err))
		writeJSONError(w, http.StatusNotFound, err)
		return
	}

	logging.AddLogAttr(r.Context(), slog.String("article_title", article.Title))

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(article)
}

func (h *handlers) handleDeleteArticle(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountID(r.Context())
	articleID := chi.URLParam(r, "id")

	logging.AddLogAttr(r.Context(), slog.String("article_id", articleID))

	result, err := h.service.DeleteArticle(r.Context(), accountID, articleID)
	if err != nil {
		handleServiceError(w, r, err, "delete article")
		return
	}

	logging.AddLogAttr(r.Context(), slog.Int("deleted", result.Deleted))

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(deleteArticleResponse{Deleted: result.Deleted})
}

func (h *handlers) handleDeleteAllArticles(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountID(r.Context())

	result, err := h.service.DeleteAllArticles(r.Context(), accountID)
	if err != nil {
		handleServiceError(w, r, err, "delete all articles")
		return
	}

	logging.AddLogAttr(r.Context(), slog.Int("deleted", result.Deleted))

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(deleteArticleResponse{Deleted: result.Deleted})
}

func (h *handlers) handleToggleFavorite(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountID(r.Context())
	articleID := chi.URLParam(r, "id")

	logging.AddLogAttr(r.Context(), slog.String("article_id", articleID))

	newStatus, err := h.service.ToggleFavorite(r.Context(), accountID, articleID)
	if err != nil {
		handleServiceError(w, r, err, "toggle favorite")
		return
	}

	logging.AddLogAttr(r.Context(), slog.Bool("favorite", newStatus))

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(favoriteResponse{Favorite: newStatus})
}

func (h *handlers) handleSendArticle(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountID(r.Context())
	articleID := chi.URLParam(r, "id")

	logging.AddLogAttr(r.Context(), slog.String("article_id", articleID))

	article, err := h.service.GetArticle(r.Context(), accountID, articleID)
	if err != nil {
		logging.AddRequestError(r.Context(), fmt.Errorf("db error: %w", err))
		writeJSONError(w, http.StatusNotFound, err)
		return
	}

	logging.AddLogAttr(r.Context(), slog.String("article_title", article.Title))

	epubBytes, err := h.service.GenerateEPUB(article)
	if err != nil {
		handleServiceError(w, r, err, "generate epub")
		return
	}

	deviceEmail, _, getErr := h.service.GetUserDeviceEmail(r.Context(), accountID)
	if getErr != nil {
		handleServiceError(w, r, getErr, "get user device email")
		return
	}
	if deviceEmail == "" {
		writeJSONError(w, http.StatusBadRequest, errors.New("user device email not configured"))
		return
	}

	logging.AddLogAttr(r.Context(), slog.String("destination_email", deviceEmail))

	emailResp, err := h.service.SendArticle(r.Context(), deviceEmail, epubBytes)
	if err != nil {
		handleServiceError(w, r, err, "send article")
		return
	}

	if dbErr := h.service.GetDBError(); dbErr != nil {
		logging.AddRequestError(r.Context(), fmt.Errorf("db error: %w", dbErr))
	}

	if emailResp != nil {
		logging.AddLogAttr(r.Context(), slog.String("message_id", emailResp.MessageID))
	}

	w.WriteHeader(http.StatusOK)

	if auth.HasSendsCount(r.Context()) {
		sendsCount := auth.GetSendsCount(r.Context())
		_ = json.NewEncoder(w).Encode(sendArticleResponseWithCount{
			Status:     "sent",
			SendsCount: sendsCount + 1,
		})
		return
	}

	_ = json.NewEncoder(w).Encode(sendArticleResponse{
		Status: "sent",
	})
}
