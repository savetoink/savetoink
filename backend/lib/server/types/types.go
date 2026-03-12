// Package types provides common HTTP request/response types for the API.
package types

import (
	"time"

	"github.com/shaftoe/savetoink/backend/lib/model"
)

// ArticleRequest represents a request to create an article.
type ArticleRequest struct {
	URL string `json:"url"`

	// Whether to send the article after processing is complete.
	SendOnComplete bool `json:"send_on_complete"`
}

// ArticleResponse represents a response for article creation.
type ArticleResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// HealthResponse represents a health check response.
type HealthResponse struct {
	Status string `json:"status"`
}

// ListArticlesResponse represents a response for listing articles.
type ListArticlesResponse struct {
	Articles []*model.Article `json:"articles"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Total    int              `json:"total"`
	HasMore  bool             `json:"has_more"`
}

// DeleteArticleResponse represents a response for deleting an article.
type DeleteArticleResponse struct {
	Deleted int `json:"deleted"`
}

// FavoriteResponse represents a response for toggling favorite status.
type FavoriteResponse struct {
	Favorite bool `json:"favorite"`
}

// SendArticleResponse represents a response for sending an article.
type SendArticleResponse struct {
	Status string `json:"status"`
}

// SendArticleResponseWithCount represents a send article response with sends count.
type SendArticleResponseWithCount struct {
	Status     string `json:"status"`
	SendsCount int    `json:"sends_count"`
}

// AuthTokenExchangeRequest represents a request to exchange an authorization code for a token.
type AuthTokenExchangeRequest struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
	GrantType   string `json:"grant_type"`
}

// AuthTokenExchangeResponse represents a response from token exchange.
type AuthTokenExchangeResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Email        string `json:"email,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// DeviceRequest represents a request to set device email.
type DeviceRequest struct {
	DeviceEmail string `json:"device_email"`
	AutoSend    bool   `json:"auto_send"`
}

// DeviceResponse represents a response for device email operations.
type DeviceResponse struct {
	DeviceEmail string `json:"device_email"`
	AutoSend    bool   `json:"auto_send"`
}

// UserProfileResponse represents a user profile response.
type UserProfileResponse struct {
	Account     string `json:"account"`
	Email       string `json:"email"`
	DeviceEmail string `json:"device_email"`
	AutoSend    bool   `json:"auto_send"`
}

// SendsResponse represents a sends quota status response.
type SendsResponse struct {
	TotalSends        int       `json:"total_sends"`
	CurrentSends      int       `json:"current_sends"`
	MaxSendsPerPeriod int       `json:"max_sends_per_period"`
	PeriodDays        int       `json:"period_days"`
	RemainingSends    int       `json:"remaining_sends"`
	PeriodResetDate   time.Time `json:"period_reset_date"`
}

type SendsResponseNoLimits struct {
	TotalSends int `json:"total_sends"`
}
