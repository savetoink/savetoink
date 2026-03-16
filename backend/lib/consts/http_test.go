package consts

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHTTPServer_Constants(t *testing.T) {
	tests := []struct {
		name  string
		value time.Duration
	}{
		{"ReadTimeout", ReadTimeout},
		{"WriteTimeout", WriteTimeout},
		{"IdleTimeout", IdleTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Greater(t, tt.value, time.Duration(0), "HTTP timeout constant should be positive")
		})
	}
}

func TestHTTPServer_TimeoutOrdering(t *testing.T) {
	assert.LessOrEqual(t, ReadTimeout, WriteTimeout, "ReadTimeout should be <= WriteTimeout")
	assert.LessOrEqual(t, WriteTimeout, IdleTimeout, "WriteTimeout should be <= IdleTimeout")
}

func TestHTTPServer_ExpectedValues(t *testing.T) {
	assert.Equal(t, 5*time.Second, ReadTimeout, "ReadTimeout should be 5 seconds")
	assert.Equal(t, 10*time.Second, WriteTimeout, "WriteTimeout should be 10 seconds")
	assert.Equal(t, 15*time.Second, IdleTimeout, "IdleTimeout should be 15 seconds")
}

func TestHTTPServer_DefaultPort(t *testing.T) {
	assert.Equal(t, 8080, DefaultHTTPPort, "DefaultHTTPPort should be 8080")
}
