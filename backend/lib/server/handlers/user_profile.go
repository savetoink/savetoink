package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/shaftoe/savetoink/backend/lib/auth"
	"github.com/shaftoe/savetoink/backend/lib/server/types"
	"github.com/shaftoe/savetoink/backend/lib/server/utils"
)

// HandleGetUserProfile handles the get user profile endpoint.
func (h *Handlers) HandleGetUserProfile(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountIDFromCtx(r.Context())

	deviceEmail, autoSend, err := h.service.GetUserDeviceEmailAndAutoSend(r.Context(), accountID)
	if err != nil {
		utils.HandleServiceError(w, r, err, "get user device email")
		return
	}

	profile, err := h.service.GetUserProfile(r.Context(), accountID)
	if err != nil {
		utils.HandleServiceError(w, r, err, "get user profile")
		return
	}

	email := ""
	if profile != nil {
		email = profile.Email
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(types.UserProfileResponse{
		Account:     accountID,
		Email:       email,
		DeviceEmail: deviceEmail,
		AutoSend:    autoSend,
	})
}

// HandleSetDevice handles the set device email endpoint.
func (h *Handlers) HandleSetDevice(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountIDFromCtx(r.Context())

	var req types.DeviceRequest
	if err := utils.DecodeAndValidateRequest(w, r, &req); err != nil {
		return
	}

	if err := h.service.SetUserDeviceEmailWithAutoSend(r.Context(), accountID, req.DeviceEmail, req.AutoSend); err != nil {
		utils.HandleServiceError(w, r, err, "set user device email")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(types.DeviceResponse(req))
}

// HandleDeleteDevice handles the delete device email endpoint.
func (h *Handlers) HandleDeleteDevice(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountIDFromCtx(r.Context())

	err := h.service.DeleteUserDeviceEmail(r.Context(), accountID)
	if err != nil {
		utils.HandleServiceError(w, r, err, "delete user device email")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(types.DeviceResponse{
		DeviceEmail: "",
		AutoSend:    false,
	})
}
