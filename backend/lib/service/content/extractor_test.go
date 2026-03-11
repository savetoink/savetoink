package content

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDOMExtractor_ExtractHTML(t *testing.T) {
	extractor := NewDOMExtractor()

	t.Run("valid html", func(t *testing.T) {
		html := `<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
<h1>Main Heading</h1>
<p>This is a paragraph.</p>
</body>
</html>`

		doc, err := extractor.Extract(context.Background(), strings.NewReader(html))
		require.NoError(t, err)
		assert.NotNil(t, doc)
	})

	t.Run("html with special characters", func(t *testing.T) {
		html := `<html>
<body>
<p>Special chars: &amp; &lt; &gt; &quot;</p>
</body>
</html>`

		doc, err := extractor.Extract(context.Background(), strings.NewReader(html))
		require.NoError(t, err)
		assert.NotNil(t, doc)
	})

	t.Run("html with unicode", func(t *testing.T) {
		html := `<html>
<body>
<p>Hello 世界 🌍</p>
</body>
</html>`

		doc, err := extractor.Extract(context.Background(), strings.NewReader(html))
		require.NoError(t, err)
		assert.NotNil(t, doc)
	})

	t.Run("malformed html", func(t *testing.T) {
		html := `<html>
<body>
<p>Unclosed paragraph
</body>
</html>`

		doc, err := extractor.Extract(context.Background(), strings.NewReader(html))
		require.NoError(t, err)
		assert.NotNil(t, doc)
	})

	t.Run("empty html", func(t *testing.T) {
		html := ""

		doc, err := extractor.Extract(context.Background(), strings.NewReader(html))
		require.NoError(t, err)
		assert.NotNil(t, doc)
	})

	t.Run("invalid html", func(t *testing.T) {
		html := "not html at all"

		doc, err := extractor.Extract(context.Background(), strings.NewReader(html))
		require.NoError(t, err)
		assert.NotNil(t, doc)
	})

	t.Run("nil reader", func(t *testing.T) {
		_, err := extractor.Extract(context.Background(), nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrHTMLParseFailed)
	})
}
