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

// HTTP server timeout constants.
const (
	// ReadTimeout is the maximum duration for reading the entire request, including the body.
	ReadTimeout = 5 * time.Second
	// WriteTimeout is the maximum duration before timing out writes of the response.
	WriteTimeout = 10 * time.Second
	// IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled.
	IdleTimeout = 15 * time.Second
)

// Auth0 client timeout constants.
const (
	// Auth0ClientTimeout is the timeout for Auth0 API calls.
	Auth0ClientTimeout = 10 * time.Second
)

// EmailProvider defines the email provider type.
type EmailProvider string

const (
	// EmailBackendMailjet indicates MailJet email backend.
	EmailBackendMailjet EmailProvider = "mailjet"
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

// Email constants.
const (
	// DefaultSubject is the default email subject.
	DefaultSubject = "Document"

	// MaxSubjectLength is the maximum length for email subjects.
	MaxSubjectLength = 100
)

// DynamoDB constants.
const (
	// DynamoDBBatchSize is the maximum number of items in a BatchWriteItem operation.
	DynamoDBBatchSize = 25

	// DynamoDBGSIName is the name of the Global Secondary Index for sorting articles by creation date (includes tags).
	DynamoDBGSIName = "AccountCreatedAtIndexWithTags"
)

// Content extraction constants.
const (
	// WordsPerMinute is the average reading speed used to calculate estimated reading time.
	WordsPerMinute = 250
)

// EPUB constants.
const (
	// DefaultChapterTitle is the default title for single-chapter EPUBs.
	DefaultChapterTitle = "Chapter 1"

	// DefaultChapterFilename is the default filename for a chapter in single-chapter EPUBs.
	DefaultChapterFilename = "chapter1.xhtml"
)

// Error messages.
const (
	// ErrInvalidArticleID is the error message for an invalid article ID.
	ErrInvalidArticleID = "invalid article id"
)
