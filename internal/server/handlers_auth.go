// Package server provides HTTP handlers and middleware for the savetoink application.
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/shaftoe/savetoink/internal/model"
)

func (h *handlers) handleAuthTokenExchange(w http.ResponseWriter, r *http.Request) {
	var req authTokenExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "failed to decode request body: " + err.Error()})
		return
	}

	if req.Code == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "missing code in request body"})
		return
	}

	if req.RedirectURI == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "missing redirect_uri in request body"})
		return
	}

	if req.GrantType == "" {
		req.GrantType = "authorization_code"
	}

	addLogAttr(r.Context(), slog.String("redirect_uri", req.RedirectURI))

	tokenReq := h.buildTokenRequest(req)
	resp, body, err := h.executeTokenExchange(tokenReq) //nolint:bodyclose // body is closed inside executeTokenExchange
	if err != nil {
		addLogAttr(r.Context(), slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "failed to exchange token with Auth0: " + err.Error()})
		return
	}

	if resp.StatusCode != http.StatusOK {
		h.handleAuth0Error(r.Context(), w, body)
		return
	}

	var tokenResp authTokenExchangeResponse
	if unmarshalErr := json.Unmarshal(body, &tokenResp); unmarshalErr != nil {
		addLogAttr(r.Context(), slog.String("error", unmarshalErr.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "failed to decode token response: " + unmarshalErr.Error()})
		return
	}

	addLogAttr(r.Context(), slog.String("token_type", tokenResp.TokenType))
	addLogAttr(r.Context(), slog.Int("expires_in", tokenResp.ExpiresIn))

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(tokenResp)
}

func (h *handlers) buildTokenRequest(req authTokenExchangeRequest) *http.Request {
	auth0TokenURL := "https://" + h.cfg.Auth0Domain + "/oauth/token"
	tokenReq := url.Values{}
	tokenReq.Set("grant_type", req.GrantType)
	tokenReq.Set("client_id", h.cfg.Auth0ClientID)
	tokenReq.Set("client_secret", h.cfg.Auth0ClientSecret)
	tokenReq.Set("code", req.Code)
	tokenReq.Set("redirect_uri", req.RedirectURI)

	httpReq, _ := http.NewRequestWithContext(
		context.Background(),
		"POST",
		auth0TokenURL,
		bytes.NewBufferString(tokenReq.Encode()),
	)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return httpReq
}

func (h *handlers) executeTokenExchange(httpReq *http.Request) (*http.Response, []byte, error) {
	resp, err := h.client.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute token exchange: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return resp, body, nil
}

func (h *handlers) handleAuth0Error(ctx context.Context, w http.ResponseWriter, body []byte) {
	var errResp struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	_ = json.Unmarshal(body, &errResp)
	addLogAttr(ctx, slog.String("auth0_error", errResp.Error))
	w.WriteHeader(http.StatusUnauthorized)

	if errResp.ErrorDescription != "" {
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: errResp.ErrorDescription})
	} else {
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "token exchange failed: " + errResp.Error})
	}
}
