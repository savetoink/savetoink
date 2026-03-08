// Package content provides article extraction functionality from web pages.
package content

// ProcessArticleEvent is the payload for processing an article.
type ProcessArticleEvent struct {
	RequestID      string           `json:"request_id"`
	URL            string           `json:"url"`
	ArticleID      string           `json:"article_id"`
	AccountID      string           `json:"account_id"`
	InheritedAttrs []map[string]any `json:"inherited_attrs"`
}
