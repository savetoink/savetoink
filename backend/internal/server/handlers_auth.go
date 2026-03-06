// Package server provides HTTP handlers and middleware for the savetoink application.
package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/shaftoe/savetoink/backend/internal/model"
)

const (
	jwtPartsCount = 3
)

//nolint:funlen // handler function is naturally longer due to Auth0 token exchange logic
func (h *handlers) handleAuthTokenExchange(w http.ResponseWriter, r *http.Request) {
	var req authTokenExchangeRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "failed to decode request body: " + decodeErr.Error()})
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
	//nolint:bodyclose // response is closed in executeTokenExchange
	resp, body, execErr := h.executeTokenExchange(tokenReq)
	if execErr != nil {
		addRequestError(r.Context(), execErr)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "failed to exchange token with Auth0: " + execErr.Error()})
		return
	}

	if resp.StatusCode != http.StatusOK {
		h.handleAuth0Error(r.Context(), w, body)
		return
	}

	var tokenResp authTokenExchangeResponse
	if unmarshalErr := json.Unmarshal(body, &tokenResp); unmarshalErr != nil {
		addRequestError(r.Context(), unmarshalErr)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "failed to decode token response: " + unmarshalErr.Error()})
		return
	}

	addLogAttr(r.Context(), slog.String("token_type", tokenResp.TokenType))
	addLogAttr(r.Context(), slog.Int("expires_in", tokenResp.ExpiresIn))

	if tokenResp.IDToken != "" {
		email, extractErr := extractEmailFromIDToken(tokenResp.IDToken)
		if extractErr != nil {
			addRequestError(r.Context(), fmt.Errorf("failed to extract email from id_token: %w", extractErr))
		} else {
			tokenResp.Email = email
			if storeErr := h.storeUserEmail(r.Context(), email, tokenResp.AccessToken); storeErr != nil {
				addRequestError(r.Context(), fmt.Errorf("failed to store user email: %w", storeErr))
			}
		}
	}

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
	resp, err := h.client.Do(httpReq) //nolint:gosec // HTTP request to Auth0 is from config, not user input
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute token exchange: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", readErr)
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

func extractEmailFromIDToken(idToken string) (string, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != jwtPartsCount {
		return "", errors.New("invalid id_token format")
	}

	payload := parts[1]
	decoded, decodeErr := base64.RawURLEncoding.DecodeString(payload)
	if decodeErr != nil {
		return "", fmt.Errorf("failed to decode id_token payload: %w", decodeErr)
	}

	var claims struct {
		Email string `json:"email"`
	}
	if unmarshalErr := json.Unmarshal(decoded, &claims); unmarshalErr != nil {
		return "", fmt.Errorf("failed to parse id_token claims: %w", unmarshalErr)
	}

	if claims.Email == "" {
		return "", errors.New("email claim not found in id_token")
	}

	return claims.Email, nil
}

func (h *handlers) storeUserEmail(ctx context.Context, email, accessToken string) error {
	if h.service == nil {
		return errors.New("service not configured")
	}

	accountID, subjectErr := getSubjectFromIDToken(accessToken)
	if subjectErr != nil {
		return fmt.Errorf("failed to extract subject from access token: %w", subjectErr)
	}

	if setErr := h.service.SetUserEmail(ctx, accountID, email); setErr != nil {
		return fmt.Errorf("failed to set user email: %w", setErr)
	}

	return nil
}

func getSubjectFromIDToken(idToken string) (string, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != jwtPartsCount {
		return "", errors.New("invalid id_token format")
	}

	payload := parts[1]
	decoded, decodeErr := base64.RawURLEncoding.DecodeString(payload)
	if decodeErr != nil {
		return "", fmt.Errorf("failed to decode id_token payload: %w", decodeErr)
	}

	var claims struct {
		Subject string `json:"sub"`
	}
	if unmarshalErr := json.Unmarshal(decoded, &claims); unmarshalErr != nil {
		return "", fmt.Errorf("failed to parse id_token claims: %w", unmarshalErr)
	}

	if claims.Subject == "" {
		return "", errors.New("subject claim not found in id_token")
	}

	return claims.Subject, nil
}
