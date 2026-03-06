package server

import (
	"encoding/json"
	"net/http"

	"github.com/shaftoe/savetoink/backend/internal/server/auth"
)

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
		handleServiceError(w, r, err, "get user device email")
		return
	}

	profile, err := h.service.GetUserProfile(r.Context(), accountID)
	if err != nil {
		handleServiceError(w, r, err, "get user profile")
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

func (h *handlers) handleSetDevice(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountID(r.Context())

	var req deviceRequest
	if err := decodeAndValidateRequest(w, r, &req); err != nil {
		return
	}

	if err := h.service.SetUserDeviceEmailWithAutoSend(r.Context(), accountID, req.DeviceEmail, req.AutoSend); err != nil {
		handleServiceError(w, r, err, "set user device email")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(deviceResponse(req))
}

func (h *handlers) handleDeleteDevice(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountID(r.Context())

	err := h.service.DeleteUserDeviceEmail(r.Context(), accountID)
	if err != nil {
		handleServiceError(w, r, err, "delete user device email")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(deviceResponse{
		DeviceEmail: "",
		AutoSend:    false,
	})
}
