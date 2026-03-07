package consts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Constants(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"ErrInvalidArticleID", ErrInvalidArticleID, "invalid article id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.value, "Error constant should not be empty")
			assert.Equal(t, tt.expected, tt.value, "Error constant should have expected value")
		})
	}
}
