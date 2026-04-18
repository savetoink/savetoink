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

	apperrors "github.com/shaftoe/savetoink/backend/lib/internal/apperrors"
	"github.com/shaftoe/savetoink/backend/lib/internal/paseto"
	"github.com/shaftoe/savetoink/backend/lib/internal/server/types"
	"github.com/shaftoe/savetoink/backend/lib/internal/server/utils"
	"github.com/shaftoe/savetoink/backend/lib/logging"
)

const (
	jwtPartsCount = 3

	// OauthTokenPath is the OAuth token endpoint path.
	OauthTokenPath = "/oauth/token" //nolint:gosec // this is an API endpoint path, not a credential
)

// HandleAuthTokenExchange handles the OAuth token exchange endpoint.
// It exchanges the authorization code with Auth0, extracts user claims,
// generates PASETO tokens, and returns them to the frontend.
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

	response, err := h.exchangeAndGenerateTokens(r, req)
	if err != nil {
		utils.HandleServiceError(w, r, err, "exchange auth tokens")
		return
	}

	logging.AddLogAttr(r.Context(), slog.String("token_type", "Bearer"))
	logging.AddLogAttr(r.Context(), slog.Int("access_expires_in", response.AccessExpiresIn))

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response) //nolint:gosec // returning PASETO access token, not a secret
}

func (h *Handlers) exchangeAndGenerateTokens(
	r *http.Request, req types.AuthTokenExchangeRequest,
) (*types.AuthTokenExchangeResponse, error) {
	tokenReq := h.buildTokenRequest(req)
	//nolint:bodyclose // response is closed in executeTokenExchange
	resp, body, execErr := h.executeTokenExchange(tokenReq)
	if execErr != nil {
		return nil, fmt.Errorf("failed to exchange token with Auth0: %w", execErr)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, h.auth0Error(r.Context(), body)
	}

	var auth0Resp types.AuthTokenExchangeResponse
	if unmarshalErr := json.Unmarshal(body, &auth0Resp); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", unmarshalErr)
	}

	email, accountID := h.extractAuth0Claims(r, &auth0Resp)

	if email != "" && accountID != "" {
		if storeErr := h.storeUserEmail(r.Context(), email, auth0Resp.AccessToken); storeErr != nil {
			logging.AddRequestError(r.Context(), fmt.Errorf("failed to store user email: %w", storeErr))
		}
	}

	pasetoPair, pasetoErr := h.generatePASETOTokens(accountID, email)
	if pasetoErr != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", pasetoErr)
	}

	return &types.AuthTokenExchangeResponse{
		AccessToken:      pasetoPair.AccessToken,
		RefreshToken:     pasetoPair.RefreshToken,
		TokenType:        "Bearer",
		AccessExpiresIn:  int(pasetoPair.AccessExpiresIn),
		RefreshExpiresIn: int(pasetoPair.RefreshExpiresIn),
		Email:            email,
	}, nil
}

func (h *Handlers) extractAuth0Claims(
	r *http.Request, auth0Resp *types.AuthTokenExchangeResponse,
) (email, accountID string) {
	if auth0Resp.IDToken != "" {
		extracted, extractErr := extractEmailFromIDToken(auth0Resp.IDToken)
		if extractErr != nil {
			logging.AddRequestError(r.Context(), fmt.Errorf("failed to extract email from id_token: %w", extractErr))
		} else {
			email = extracted
		}
	}

	if auth0Resp.AccessToken != "" {
		subject, subjectErr := getSubjectFromIDToken(auth0Resp.AccessToken)
		if subjectErr != nil {
			logging.AddRequestError(r.Context(), fmt.Errorf("failed to extract subject from access token: %w", subjectErr))
		} else {
			accountID = subject
		}
	}
	return email, accountID
}

func (h *Handlers) generatePASETOTokens(accountID, email string) (*paseto.TokenPair, error) {
	if h.pasetoKeyStore == nil {
		return nil, errors.New("paseto key store not configured")
	}

	claims := paseto.TokenClaims{
		Subject: accountID,
		Email:   email,
	}

	pair, err := h.pasetoKeyStore.GenerateTokenPair(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate paseto token pair: %w", err)
	}
	return pair, nil
}

// HandleAuthRefresh handles the token refresh endpoint.
// It validates the refresh token and returns a new token pair.
func (h *Handlers) HandleAuthRefresh(w http.ResponseWriter, r *http.Request) {
	if h.pasetoKeyStore == nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, errors.New("paseto key store not configured"))
		return
	}

	var req types.AuthTokenRefreshRequest
	if decodeErr := utils.DecodeAndValidateRequest(w, r, &req); decodeErr != nil {
		return
	}

	if req.RefreshToken == "" {
		utils.WriteJSONError(w, http.StatusBadRequest, errors.New("missing refresh_token in request body"))
		return
	}

	pair, err := h.pasetoKeyStore.RefreshTokens(req.RefreshToken)
	if err != nil {
		utils.WriteJSONError(w, http.StatusUnauthorized, err)
		return
	}

	response := &types.AuthTokenExchangeResponse{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		TokenType:        "Bearer",
		AccessExpiresIn:  int(pair.AccessExpiresIn),
		RefreshExpiresIn: int(pair.RefreshExpiresIn),
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response) //nolint:gosec // returning PASETO access token, not a secret
}

func (h *Handlers) auth0Error(ctx context.Context, body []byte) error {
	var errResp struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	_ = json.Unmarshal(body, &errResp)
	logging.AddLogAttr(ctx, slog.String("auth0_error", errResp.Error))

	if errResp.ErrorDescription != "" {
		return fmt.Errorf("%s: %w", errResp.ErrorDescription, apperrors.ErrUnauthorized)
	}
	return fmt.Errorf("%s: %w", errResp.Error, apperrors.ErrUnauthorized)
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
