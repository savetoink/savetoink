package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/shaftoe/savetoink/internal/auth"
	"github.com/shaftoe/savetoink/internal/model"
)

type userProfileRequest struct {
	KindleEmail string `json:"kindle_email"`
}

type userProfileResponse struct {
	Account     string `json:"account"`
	KindleEmail string `json:"kindle_email"`
}

func (h *handlers) handleGetUserProfile(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountID(r.Context())

	kindleEmail, err := h.service.GetUserKindleEmail(r.Context(), accountID)
	if err != nil {
		addLogAttr(r.Context(), slog.String("db_error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(userProfileResponse{
		Account:     accountID,
		KindleEmail: kindleEmail,
	})
}

func (h *handlers) handleSetUserProfile(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountID(r.Context())

	var req userProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "failed to decode request body: " + err.Error()})
		return
	}

	if req.KindleEmail == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "missing kindleEmail in request body"})
		return
	}

	err := h.service.SetUserKindleEmail(r.Context(), accountID, req.KindleEmail)
	if err != nil {
		addLogAttr(r.Context(), slog.String("db_error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(userProfileResponse{
		Account:     accountID,
		KindleEmail: req.KindleEmail,
	})
}

func (h *handlers) handleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountID(r.Context())

	err := h.service.DeleteUserProfile(r.Context(), accountID)
	if err != nil {
		addLogAttr(r.Context(), slog.String("db_error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
}
