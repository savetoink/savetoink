// Package server provides HTTP handlers and middleware for the savetoink application.
package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/shaftoe/savetoink/internal/consts"
	"github.com/shaftoe/savetoink/internal/model"
	"github.com/shaftoe/savetoink/internal/server/auth"
)

func (h *handlers) handleCreateArticle(w http.ResponseWriter, r *http.Request) {
	var req articleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "failed to decode request body: " + err.Error()})
		return
	}

	if req.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "missing URL in request body"})
		return
	}

	addLogAttr(r.Context(), slog.String("url", req.URL))

	result, err := h.service.CreateArticle(r.Context(), req.URL, auth.GetAccountID(r.Context()))
	if err != nil {
		addLogAttr(r.Context(), slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
		return
	}

	addLogAttr(r.Context(), slog.String("article_id", result.Article.ID))
	addLogAttr(r.Context(), slog.String("article_url", result.Article.URL))
	addLogAttr(r.Context(), slog.String("message", result.Message))

	if result.Article.DeliveryStatus != "" {
		addLogAttr(r.Context(), slog.String("delivery_status", string(result.Article.DeliveryStatus)))
		addLogAttr(r.Context(), slog.String("email_provider", string(h.cfg.EmailProvider)))
	}

	if result.EmailResp != nil && result.EmailResp.EmailUUID != "" {
		addLogAttr(r.Context(), slog.String("email_uuid", result.EmailResp.EmailUUID))
	}

	if dbErr := h.service.GetDBError(); dbErr != nil {
		addLogAttr(r.Context(), slog.String("db_error", dbErr.Error()))
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(articleResponse{
		ID:             result.Article.ID,
		Title:          result.Article.Title,
		URL:            result.Article.URL,
		Message:        result.Message,
		DeliveryStatus: string(result.Article.DeliveryStatus),
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
		addLogAttr(r.Context(), slog.String("db_error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
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
		addLogAttr(r.Context(), slog.String("db_error", err.Error()))
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
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
		addLogAttr(r.Context(), slog.String("db_error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
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
		addLogAttr(r.Context(), slog.String("db_error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
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
		addLogAttr(r.Context(), slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
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
		addLogAttr(r.Context(), slog.String("db_error", err.Error()))
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
		return
	}

	addLogAttr(r.Context(), slog.String("article_title", article.Title))

	emailResp, err := h.service.SendArticle(r.Context(), article, accountID)
	if err != nil {
		addLogAttr(r.Context(), slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
		return
	}

	if emailResp != nil {
		addLogAttr(r.Context(), slog.String("email_uuid", emailResp.EmailUUID))
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(sendArticleResponse{
		Status:  "sent",
		Message: "article sent to Kindle successfully",
	})
}
