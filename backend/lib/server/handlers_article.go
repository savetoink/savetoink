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

	"github.com/go-chi/chi/v5"
	"github.com/shaftoe/savetoink/backend/lib/auth"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/shaftoe/savetoink/backend/lib/validation"
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
	reqID := logging.GetRequestID(r.Context())

	event := &content.ProcessArticleEvent{
		RequestID:      reqID,
		URL:            url,
		ArticleID:      articleID,
		AccountID:      accountID,
		InheritedAttrs: convertSlogAttrsToMap(inheritedAttrs),
	}

	h.processor.StartProcessing(context.Background(), event)
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

func convertSlogAttrsToMap(inheritedAttrs []slog.Attr) []map[string]any {
	if inheritedAttrs == nil {
		return nil
	}
	attrs := make([]map[string]any, len(inheritedAttrs))
	for i, attr := range inheritedAttrs {
		attrs[i] = map[string]any{attr.Key: attr.Value.Any()}
	}
	return attrs
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

	result, err := h.service.SendArticleByID(r.Context(), accountID, articleID)
	if err != nil {
		handleServiceError(w, r, err, "send article")
		return
	}

	logging.AddLogAttr(r.Context(), slog.String("article_title", result.Article.Title))
	logging.AddLogAttr(r.Context(), slog.String("destination_email", result.DeviceEmail))
	if result.EmailResp != nil {
		logging.AddLogAttr(r.Context(), slog.String("message_id", result.EmailResp.MessageID))
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
