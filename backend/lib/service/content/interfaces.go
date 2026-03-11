package content

import (
	"context"
	"io"

	"golang.org/x/net/html"

	"net/url"

	"github.com/shaftoe/savetoink/backend/lib/model"
)

// Extractor converts HTML reader to html.Node.
type Extractor interface {
	Extract(ctx context.Context, htmlReader io.Reader) (*html.Node, error)
}

// Cleaner extracts article content from html.Node.
type Cleaner interface {
	Clean(ctx context.Context, doc, node *html.Node, u *url.URL) (*model.Article, error)
}
