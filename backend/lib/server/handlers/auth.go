// Package handlers provides HTTP handlers for the savetoink application.
package handlers

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

	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/server/types"
	"github.com/shaftoe/savetoink/backend/lib/server/utils"
)

const (
	jwtPartsCount = 3

	// OauthTokenPath is the OAuth token endpoint path.
	OauthTokenPath = "/oauth/token" //nolint:gosec // this is an API endpoint path, not a credential
)

// HandleAuthTokenExchange handles the OAuth token exchange endpoint.
func (h *Handlers) HandleAuthTokenExchange(w http.ResponseWriter, r *http.Request) {
	var req types.AuthTokenExchangeRequest
	if decodeErr := utils.DecodeAndValidateRequest(w, r, &req); decodeErr != nil {
		return
	}

	if req.Code == "" {
		utils.WriteJSONError(w, http.StatusBadRequest, errors.New("missing code in request body"))
		return
	}

	if req.RedirectURI == "" {
		utils.WriteJSONError(w, http.StatusBadRequest, errors.New("missing redirect_uri in request body"))
		return
	}

	if req.GrantType == "" {
		req.GrantType = "authorization_code"
	}

	logging.AddLogAttr(r.Context(), slog.String("redirect_uri", req.RedirectURI))

	tokenReq := h.buildTokenRequest(req)
	//nolint:bodyclose // response is closed in executeTokenExchange
	resp, body, execErr := h.executeTokenExchange(tokenReq)
	if execErr != nil {
		logging.AddRequestError(r.Context(), execErr)
		utils.WriteJSONError(w, http.StatusInternalServerError,
			fmt.Errorf("failed to exchange token with Auth0: %w", execErr))
		return
	}

	if resp.StatusCode != http.StatusOK {
		h.handleAuth0Error(r.Context(), w, body)
		return
	}

	var tokenResp types.AuthTokenExchangeResponse
	if unmarshalErr := json.Unmarshal(body, &tokenResp); unmarshalErr != nil {
		logging.AddRequestError(r.Context(), unmarshalErr)
		utils.WriteJSONError(w, http.StatusInternalServerError,
			fmt.Errorf("failed to decode token response: %w", unmarshalErr))
		return
	}

	logging.AddLogAttr(r.Context(), slog.String("token_type", tokenResp.TokenType))
	logging.AddLogAttr(r.Context(), slog.Int("expires_in", tokenResp.ExpiresIn))

	if tokenResp.IDToken != "" {
		email, extractErr := extractEmailFromIDToken(tokenResp.IDToken)
		if extractErr != nil {
			logging.AddRequestError(r.Context(), fmt.Errorf("failed to extract email from id_token: %w", extractErr))
		} else {
			tokenResp.Email = email
			if storeErr := h.storeUserEmail(r.Context(), email, tokenResp.AccessToken); storeErr != nil {
				logging.AddRequestError(r.Context(), fmt.Errorf("failed to store user email: %w", storeErr))
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(tokenResp) //nolint:gosec // returning OAuth access token, not a secret
}

func (h *Handlers) buildTokenRequest(req types.AuthTokenExchangeRequest) *http.Request {
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

func (h *Handlers) executeTokenExchange(httpReq *http.Request) (*http.Response, []byte, error) {
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

func (h *Handlers) handleAuth0Error(ctx context.Context, w http.ResponseWriter, body []byte) {
	var errResp struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	_ = json.Unmarshal(body, &errResp)
	logging.AddLogAttr(ctx, slog.String("auth0_error", errResp.Error))

	if errResp.ErrorDescription != "" {
		utils.WriteJSONError(w, http.StatusUnauthorized, errors.New(errResp.ErrorDescription))
	} else {
		utils.WriteJSONError(w, http.StatusUnauthorized, fmt.Errorf("token exchange failed: %s", errResp.Error))
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

func (h *Handlers) storeUserEmail(ctx context.Context, email, accessToken string) error {
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
