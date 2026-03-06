package server

import (
	"log/slog"
	"net/http"

	"github.com/shaftoe/savetoink/backend/internal/config"
	"github.com/shaftoe/savetoink/backend/internal/model"
	"github.com/shaftoe/savetoink/backend/internal/service"
)

type articleRequest struct {
	URL string `json:"url"`
}

type articleResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type healthResponse struct {
	Status string `json:"status"`
}

type listArticlesResponse struct {
	Articles []*model.Article `json:"articles"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Total    int              `json:"total"`
	HasMore  bool             `json:"has_more"`
}

type deleteArticleResponse struct {
	Deleted int `json:"deleted"`
}

type favoriteResponse struct {
	Favorite bool `json:"favorite"`
}

type sendArticleResponse struct {
	Status string `json:"status"`
}

type sendArticleResponseWithCount struct {
	Status     string `json:"status"`
	SendsCount int    `json:"sends_count"`
}

type authTokenExchangeRequest struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
	GrantType   string `json:"grant_type"`
}

type authTokenExchangeResponse struct {
	AccessToken  string `json:"access_token"`            //nolint:gosec // OAuth2 access token, not a secret
	RefreshToken string `json:"refresh_token,omitempty"` //nolint:gosec // OAuth2 refresh token, not a secret
	IDToken      string `json:"id_token,omitempty"`
	Email        string `json:"email,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type deviceRequest struct {
	DeviceEmail string `json:"device_email"`
	AutoSend    bool   `json:"auto_send"`
}

type deviceResponse struct {
	DeviceEmail string `json:"device_email"`
	AutoSend    bool   `json:"auto_send"`
}

type handlers struct {
	cfg     *config.Config
	service service.Interface
	client  *http.Client
}

type contextKey string

const (
	requestErrorKey contextKey = "request_error"
)

type logRecord struct {
	*slog.Record
}
