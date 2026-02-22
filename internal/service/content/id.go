package content

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// CleanURL strips query parameters and fragments from a URL to ensure
// the same base URL always produces a consistent result.
// Returns the cleaned URL with scheme, host, and path only.
func CleanURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("url must be valid: %w", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", errors.New("url must have scheme and host")
	}

	path := strings.TrimSuffix(parsedURL.Path, "/")
	if path == "" {
		path = "/"
	}

	return fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, path), nil
}

// ArticleIDFromURL generates a deterministic UUID v5 for an article from its URL.
// Strips query parameters and fragments before hashing to ensure
// the same base URL always produces the same ID.
//
// Uses UUID v5 with the URL namespace as defined in RFC 4122.
func ArticleIDFromURL(rawURL string) (string, error) {
	cleanURL, err := CleanURL(rawURL)
	if err != nil {
		return "", err
	}

	id := uuid.NewSHA1(uuid.NameSpaceURL, []byte(cleanURL))

	return id.String(), nil
}
