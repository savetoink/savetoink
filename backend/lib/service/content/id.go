package content

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// CleanURL cleans a URL by removing trailing slashes and fragments while preserving query parameters.
func CleanURL(u *url.URL) string {
	path := strings.TrimSuffix(u.Path, "/")
	if path == "" {
		path = "/"
	}

	cleanURL := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, path)
	if u.RawQuery != "" {
		cleanURL += "?" + u.RawQuery
	}

	return cleanURL
}

// ArticleIDFromURL generates a deterministic UUID v5 for an article from its URL.
// Uses UUID v5 with URL namespace as defined in RFC 4122.
func ArticleIDFromURL(u *url.URL) (string, error) {
	cleanURL := CleanURL(u)

	id := uuid.NewSHA1(uuid.NameSpaceURL, []byte(cleanURL))

	return id.String(), nil
}
