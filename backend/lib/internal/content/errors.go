package content

import "errors"

var (
	// ErrInvalidURL is returned when a URL is invalid.
	ErrInvalidURL = errors.New("invalid url")
	// ErrHTMLParseFailed is returned when parsing HTML fails.
	ErrHTMLParseFailed = errors.New("failed to parse html")
	// ErrExtractionFailed is returned when content extraction fails.
	ErrExtractionFailed = errors.New("failed to extract content")
	// ErrNoContentExtracted is returned when no content is extracted.
	ErrNoContentExtracted = errors.New("no content extracted")
	// ErrNilContentNode is returned when the content node is nil.
	ErrNilContentNode = errors.New("content node is nil")
	// ErrArticleBuildFailed is returned when building an article fails.
	ErrArticleBuildFailed = errors.New("failed to build article")
)
