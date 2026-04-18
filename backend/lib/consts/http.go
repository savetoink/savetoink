package consts

import (
	"time"
)

// HTTP server timeout constants.
const (
	// ReadTimeout is the maximum duration for reading the entire request, including the body.
	ReadTimeout = 5 * time.Second
	// WriteTimeout is the maximum duration before timing out writes of the response.
	WriteTimeout = 10 * time.Second
	// IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled.
	IdleTimeout = 15 * time.Second

	// DefaultHTTPPort is the default HTTP port for the standalon backend process..
	DefaultHTTPPort = 8080
)

// RequestIDFormat is the generated request ID format (YYYYMMDD-HHMMSS.mmm).
const RequestIDFormat = "20060102-150405.000"

// BrowserlessContentURL is the URL for the Browserless content API.
const BrowserlessContentURL = "https://production-sfo.browserless.io/content"

// HTMLErrorPatterns contains patterns that indicate invalid HTML content or error pages.
var HTMLErrorPatterns = []string{
	"This website is using a security service to protect itself from online attacks. " +
		"The action you just performed triggered the security solution.",
}

// ArticleProcessingTimeout is the maximum duration for processing an article.
const ArticleProcessingTimeout = 30 * time.Second

// Auth tokens constants.
const (
	// TokenPrefix is the prefix of v4.local PASETO tokens.
	TokenPrefix = "v4.local."

	// keySize is the required size in bytes for a v4 symmetric key.
	KeySize = 32

	// DefaultAccessTTL is the default time-to-live for access tokens.
	DefaultAccessTTL = 10 * time.Second

	// DefaultRefreshTTL is the default time-to-live for refresh tokens.
	DefaultRefreshTTL = 30 * 24 * time.Hour
)
