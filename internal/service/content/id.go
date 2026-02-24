package content

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// CleanURL cleans a URL by removing trailing slashes and fragments while preserving query parameters.
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

	cleanURL := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, path)
	if parsedURL.RawQuery != "" {
		cleanURL += "?" + parsedURL.RawQuery
	}

	return cleanURL, nil
}

// ArticleIDFromURL generates a deterministic UUID v5 for an article from its URL.
// Uses UUID v5 with the URL namespace as defined in RFC 4122.
func ArticleIDFromURL(rawURL string) (string, error) {
	cleanURL, err := CleanURL(rawURL)
	if err != nil {
		return "", err
	}

	id := uuid.NewSHA1(uuid.NameSpaceURL, []byte(cleanURL))

	return id.String(), nil
}
