// Package consts provides shared constants used across the savetoink application.
package consts

import "time"

var version = "0.0.0"

// Version returns the current version of the savetoink application, to be overridden by the build process.
func Version() *string {
	return &version
}

// RunMode defines the application execution mode.
type RunMode string

const (
	// ModeCLI indicates CLI execution mode.
	ModeCLI RunMode = "cli"
	// ModeServer indicates server execution mode.
	ModeServer RunMode = "server"
)

// AuthBackend defines the authentication backend type.
type AuthBackend string

const (
	// AuthBackendSharedAPIKey indicates shared API key authentication.
	AuthBackendSharedAPIKey AuthBackend = "shared_api_key"
	// AuthBackendAuth0 indicates Auth0 JWT authentication.
	AuthBackendAuth0 AuthBackend = "auth0"
)

// Status represents the delivery status of an article.
type Status string

const (
	// StatusPending indicates that the article is pending delivery.
	StatusPending Status = "pending"
	// StatusDelivered indicates that the article has been successfully delivered.
	StatusDelivered Status = "delivered"
	// StatusFailed indicates that the article delivery has failed.
	StatusFailed Status = "failed"
)

// Pagination constants.
const (
	// MinPage is the minimum valid page number for pagination.
	MinPage = 1

	// DefaultPage is the default page number for pagination.
	DefaultPage = 1

	// MinPageSize is the minimum number of items per page.
	MinPageSize = 1

	// DefaultPageSize is the default number of items per page.
	DefaultPageSize = 20

	// MaxPageSize is the maximum number of items per page.
	MaxPageSize = 20
)

// Content extraction constants.
const (
	// WordsPerMinute is the average reading speed used to calculate estimated reading time.
	WordsPerMinute = 250

	// MinimumExtractedSize is the minimum number of characters required for extracted content.
	// Set to 0 to allow articles of any size including empty content.
	MinimumExtractedSize = 0

	// MinimumOutputSize is the minimum number of characters required in final output.
	MinimumOutputSize = 0
)

// Auth0 client timeout constants.
const (
	// Auth0ClientTimeout is the timeout for Auth0 API calls.
	Auth0ClientTimeout = 10 * time.Second
)
