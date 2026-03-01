package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/shaftoe/savetoink/internal/model"
	"github.com/shaftoe/savetoink/internal/server/auth"
)

type userProfileRequest struct {
	DeviceEmail string `json:"device_email"`
	AutoSend    bool   `json:"auto_send"`
}

type userProfileResponse struct {
	Account     string `json:"account"`
	Email       string `json:"email"`
	DeviceEmail string `json:"device_email"`
	AutoSend    bool   `json:"auto_send"`
}

func (h *handlers) handleGetUserProfile(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountID(r.Context())

	deviceEmail, autoSend, err := h.service.GetUserDeviceEmail(r.Context(), accountID)
	if err != nil {
		addLogAttr(r.Context(), slog.String("db_error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
		return
	}

	profile, err := h.service.GetUserProfile(r.Context(), accountID)
	if err != nil {
		addLogAttr(r.Context(), slog.String("db_error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
		return
	}

	email := ""
	if profile != nil {
		email = profile.Email
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(userProfileResponse{
		Account:     accountID,
		Email:       email,
		DeviceEmail: deviceEmail,
		AutoSend:    autoSend,
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

	if req.DeviceEmail == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "missing device_email in request body"})
		return
	}

	if err := h.service.SetUserDeviceEmailWithAutoSend(r.Context(), accountID, req.DeviceEmail, req.AutoSend); err != nil {
		addLogAttr(r.Context(), slog.String("db_error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
		return
	}

	profile, err := h.service.GetUserProfile(r.Context(), accountID)
	if err != nil {
		addLogAttr(r.Context(), slog.String("db_error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
		return
	}

	email := ""
	if profile != nil {
		email = profile.Email
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(userProfileResponse{
		Account:     accountID,
		Email:       email,
		DeviceEmail: req.DeviceEmail,
		AutoSend:    req.AutoSend,
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
