// Package server provides HTTP handlers and middleware for the savetoink application.
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/shaftoe/savetoink/backend/internal/consts"
	"github.com/shaftoe/savetoink/backend/internal/server/auth"
	"github.com/shaftoe/savetoink/backend/internal/validation"
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

	addLogAttr(r.Context(), slog.String("url", req.URL))

	article, err := h.service.CreateArticle(r.Context(), req.URL, auth.GetAccountID(r.Context()))
	if err != nil {
		handleServiceError(w, r, err, "create article")
		return
	}

	addLogAttr(r.Context(), slog.String("article_id", article.ID))

	if dbErr := h.service.GetDBError(); dbErr != nil {
		addRequestError(r.Context(), fmt.Errorf("db error: %w", dbErr))
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(articleResponse{
		ID:    article.ID,
		Title: article.Title,
		URL:   article.URL,
	})
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

	addLogAttr(r.Context(), slog.Int("page", page))
	addLogAttr(r.Context(), slog.Int("page_size", pageSize))
	addLogAttr(r.Context(), slog.Int("total", result.Total))

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

	addLogAttr(r.Context(), slog.String("article_id", articleID))

	article, err := h.service.GetArticle(r.Context(), accountID, articleID)
	if err != nil {
		addRequestError(r.Context(), fmt.Errorf("db error: %w", err))
		writeJSONError(w, http.StatusNotFound, err)
		return
	}

	addLogAttr(r.Context(), slog.String("article_title", article.Title))

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(article)
}

func (h *handlers) handleDeleteArticle(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountID(r.Context())
	articleID := chi.URLParam(r, "id")

	addLogAttr(r.Context(), slog.String("article_id", articleID))

	result, err := h.service.DeleteArticle(r.Context(), accountID, articleID)
	if err != nil {
		handleServiceError(w, r, err, "delete article")
		return
	}

	addLogAttr(r.Context(), slog.Int("deleted", result.Deleted))

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

	addLogAttr(r.Context(), slog.Int("deleted", result.Deleted))

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(deleteArticleResponse{Deleted: result.Deleted})
}

func (h *handlers) handleToggleFavorite(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountID(r.Context())
	articleID := chi.URLParam(r, "id")

	addLogAttr(r.Context(), slog.String("article_id", articleID))

	newStatus, err := h.service.ToggleFavorite(r.Context(), accountID, articleID)
	if err != nil {
		handleServiceError(w, r, err, "toggle favorite")
		return
	}

	addLogAttr(r.Context(), slog.Bool("favorite", newStatus))

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(favoriteResponse{Favorite: newStatus})
}

func (h *handlers) handleSendArticle(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountID(r.Context())
	articleID := chi.URLParam(r, "id")

	addLogAttr(r.Context(), slog.String("article_id", articleID))

	article, err := h.service.GetArticle(r.Context(), accountID, articleID)
	if err != nil {
		addRequestError(r.Context(), fmt.Errorf("db error: %w", err))
		writeJSONError(w, http.StatusNotFound, err)
		return
	}

	addLogAttr(r.Context(), slog.String("article_title", article.Title))

	emailResp, err := h.service.SendArticle(r.Context(), article, accountID)
	if err != nil {
		handleServiceError(w, r, err, "send article")
		return
	}

	if emailResp != nil {
		addLogAttr(r.Context(), slog.String("message_id", emailResp.MessageID))
	}

	deviceEmail, _, getErr := h.service.GetUserDeviceEmail(r.Context(), accountID)
	if getErr == nil && deviceEmail != "" {
		addLogAttr(r.Context(), slog.String("destination_email", deviceEmail))
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
