// Package content provides article extraction functionality from web pages.
package content

import (
	"context"
	"fmt"
	"io"

	"github.com/go-shiori/dom"
	"golang.org/x/net/html"
)

// DOMExtractor handles parsing HTML content into a DOM tree.
type DOMExtractor struct{}

// NewDOMExtractor creates a new DOMExtractor instance.
func NewDOMExtractor() *DOMExtractor {
	return &DOMExtractor{}
}

// Extract parses HTML content and returns the DOM tree.
func (e *DOMExtractor) Extract(_ context.Context, htmlReader io.Reader) (*html.Node, error) {
	if htmlReader == nil {
		return nil, fmt.Errorf("%w: reader is nil", ErrHTMLParseFailed)
	}
	doc, err := dom.Parse(htmlReader)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrHTMLParseFailed, err)
	}
	return doc, nil
}
