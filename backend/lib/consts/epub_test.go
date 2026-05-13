package consts

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEPUB_Constants(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"DefaultChapterTitle", DefaultChapterTitle, "Chapter 1"},
		{"DefaultChapterFilename", DefaultChapterFilename, "chapter1.xhtml"},
		{"EPUBStylesheetFilename", EPUBStylesheetFilename, "styles.css"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.value, "EPUB constant should not be empty")
			assert.Equal(t, tt.expected, tt.value, "EPUB constant should have expected value")
		})
	}
}

func TestEPUBStylesheet_ContainsPreRules(t *testing.T) {
	assert.NotEmpty(t, EPUBStylesheet, "EPUBStylesheet should not be empty")

	expectedRules := []string{
		"white-space: pre-wrap",
		"font-family: monospace",
		"background-color",
		"padding",
	}

	for _, rule := range expectedRules {
		assert.True(t, strings.Contains(EPUBStylesheet, rule),
			"EPUBStylesheet should contain %q", rule)
	}
}

func TestEPUBStylesheet_HasPreAndCodeSelectors(t *testing.T) {
	assert.True(t, strings.Contains(EPUBStylesheet, "pre {"),
		"EPUBStylesheet should contain pre selector")
	assert.True(t, strings.Contains(EPUBStylesheet, "code {"),
		"EPUBStylesheet should contain code selector")
	assert.True(t, strings.Contains(EPUBStylesheet, "pre code {"),
		"EPUBStylesheet should contain pre code selector")
}
