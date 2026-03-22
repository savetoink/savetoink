// Package handlers provides HTTP handlers for the savetoink application.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/internal/auth"
	"github.com/shaftoe/savetoink/backend/lib/internal/content"
	"github.com/shaftoe/savetoink/backend/lib/internal/server/types"
	"github.com/shaftoe/savetoink/backend/lib/internal/server/utils"
	"github.com/shaftoe/savetoink/backend/lib/internal/validation"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/processor"
	"github.com/shaftoe/savetoink/backend/lib/service"
)

// Handlers manages HTTP handlers for the savetoink application.
type Handlers struct {
	cfg       *config.Config
	service   service.Interface
	client    *http.Client
	processor processor.Processor
}

// HandleCreateArticle handles the article creation endpoint.
func (h *Handlers) HandleCreateArticle(w http.ResponseWriter, r *http.Request) {
	var req types.ArticleRequest
	if err := utils.DecodeAndValidateRequest(w, r, &req); err != nil {
		return
	}

	accountID := auth.GetAccountIDFromCtx(r.Context())
	sendsCount, err := h.validateAndCheckQuota(w, r, req, accountID)
	if err != nil {
		return
	}

	u, err := h.validateURL(w, r, req)
	if err != nil {
		return
	}

	article, err := h.service.CreateArticle(r.Context(), u, accountID)
	if err != nil {
		utils.HandleServiceError(w, r, err, "create article")
		return
	}

	h.writeArticleResponse(w, article)
	h.startArticleProcessing(r.Context(), req.URL, article.ID, accountID, req.SendOnComplete, sendsCount)
}

func (h *Handlers) validateAndCheckQuota(
	w http.ResponseWriter,
	r *http.Request,
	req types.ArticleRequest,
	accountID string,
) (int, error) {
	var sendsCount int
	if req.SendOnComplete {
		if err := utils.CheckEmailBackendEnabled(w, r, h.cfg.EmailProvider); err != nil {
			return 0, fmt.Errorf("email backend check failed: %w", err)
		}

		sendsError, err := utils.CheckQuotaAndDeviceEmail(r.Context(), w, r, h.service, h.cfg.AuthBackend, accountID, h.cfg.DisableQuotaCheck)
		if err != nil {
			return 0, fmt.Errorf("quota and device email check failed: %w", err)
		}
		sendsCount = sendsError
	}

	if req.URL == "" {
		utils.WriteJSONError(w, http.StatusBadRequest, errors.New("missing URL in request body"))
		return 0, errors.New("missing URL")
	}

	return sendsCount, nil
}

func (h *Handlers) validateURL(w http.ResponseWriter, r *http.Request, req types.ArticleRequest) (*url.URL, error) {
	u, err := validation.ValidateURL(req.URL)
	if err != nil {
		urlErr := fmt.Errorf("invalid URL: %w", err)
		utils.WriteJSONError(w, http.StatusBadRequest, urlErr)
		return nil, urlErr
	}

	logging.AddLogAttr(r.Context(), slog.String("url", req.URL))
	logging.AddLogAttr(r.Context(), slog.Bool("send_on_complete", req.SendOnComplete))
	return u, nil
}

func (h *Handlers) writeArticleResponse(w http.ResponseWriter, article *model.Article) {
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(types.ArticleResponse{
		ID:    article.ID,
		Title: article.Title,
		URL:   article.URL,
	})
}

func (h *Handlers) startArticleProcessing(
	ctx context.Context,
	articleURL, articleID, accountID string,
	sendOnComplete bool,
	_ int,
) {
	inheritedAttrs := logging.ExtractInheritedLogAttrs(ctx)
	reqID := logging.GetRequestID(ctx)

	event := &content.ProcessArticleEvent{
		RequestID:      reqID,
		URL:            articleURL,
		ArticleID:      articleID,
		AccountID:      accountID,
		InheritedAttrs: logging.ConvertSlogAttrsToMap(inheritedAttrs),
		SendOnComplete: sendOnComplete,
	}

	h.processor.StartProcessing(context.Background(), event)
}

// HandleGetArticles handles the get articles list endpoint.
func (h *Handlers) HandleGetArticles(w http.ResponseWriter, r *http.Request) {
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

	accountID := auth.GetAccountIDFromCtx(r.Context())

	result, err := h.service.GetArticlesMetadata(r.Context(), accountID, page, pageSize, favoriteFilter)
	if err != nil {
		utils.HandleServiceError(w, r, err, "get articles metadata")
		return
	}

	logging.AddLogAttr(r.Context(), slog.Int("page", page))
	logging.AddLogAttr(r.Context(), slog.Int("page_size", pageSize))
	logging.AddLogAttr(r.Context(), slog.Int("total", result.Total))

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(types.ListArticlesResponse{
		Articles: result.Articles,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
		HasMore:  result.HasMore,
	})
}

// HandleGetArticle handles the get single article endpoint.
func (h *Handlers) HandleGetArticle(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountIDFromCtx(r.Context())
	articleID := utils.GetArticleID(r)

	article, err := h.service.GetArticle(r.Context(), accountID, articleID)
	if err != nil {
		logging.AddRequestError(r.Context(), fmt.Errorf("db error: %w", err))
		utils.WriteJSONError(w, http.StatusNotFound, err)
		return
	}

	logging.AddLogAttr(r.Context(), slog.String("article_title", article.Title))

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(article)
}

// HandleDeleteArticle handles the delete article endpoint.
func (h *Handlers) HandleDeleteArticle(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountIDFromCtx(r.Context())
	articleID := utils.GetArticleID(r)

	result, err := h.service.DeleteArticle(r.Context(), accountID, articleID)
	if err != nil {
		utils.HandleServiceError(w, r, err, "delete article")
		return
	}

	logging.AddLogAttr(r.Context(), slog.Int("deleted", result.Deleted))

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(types.DeleteArticleResponse{Deleted: result.Deleted})
}

// HandleToggleFavorite handles the toggle favorite endpoint.
func (h *Handlers) HandleToggleFavorite(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountIDFromCtx(r.Context())
	articleID := utils.GetArticleID(r)

	newStatus, err := h.service.ToggleFavorite(r.Context(), accountID, articleID)
	if err != nil {
		utils.HandleServiceError(w, r, err, "toggle favorite")
		return
	}

	logging.AddLogAttr(r.Context(), slog.Bool("favorite", newStatus))

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(types.FavoriteResponse{Favorite: newStatus})
}

// HandleSendArticle handles the send article endpoint.
func (h *Handlers) HandleSendArticle(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountIDFromCtx(r.Context())
	articleID := utils.GetArticleID(r)

	if err := utils.CheckEmailBackendEnabled(w, r, h.cfg.EmailProvider); err != nil {
		return
	}

	sendsCount, err := utils.CheckQuotaAndDeviceEmail(r.Context(), w, r, h.service, h.cfg.AuthBackend, accountID, h.cfg.DisableQuotaCheck)
	if err != nil {
		return
	}

	result, err := h.service.SendArticleByID(r.Context(), accountID, articleID)
	if err != nil {
		utils.HandleServiceError(w, r, err, "send article")
		return
	}

	logging.AddLogAttr(r.Context(), slog.String("article_title", result.Article.Title))
	logging.AddLogAttr(r.Context(), slog.String("destination_email", result.DeviceEmail))
	if result.EmailResp != nil {
		logging.AddLogAttr(r.Context(), slog.String("message_id", result.EmailResp.MessageID))
	}

	w.WriteHeader(http.StatusOK)

	if h.cfg.AuthBackend == consts.AuthBackendAuth0 {
		_ = json.NewEncoder(w).Encode(types.SendArticleResponseWithCount{
			Status:     "sent",
			SendsCount: sendsCount + 1,
		})
		return
	}

	_ = json.NewEncoder(w).Encode(types.SendArticleResponse{
		Status: "sent",
	})
}
