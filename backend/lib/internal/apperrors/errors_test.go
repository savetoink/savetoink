package apperrors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Constants(t *testing.T) {
	tests := []struct {
		name     string
		value    error
		expected string
	}{
		{"ErrNotFound", ErrNotFound, "not found"},
		{"ErrInvalid", ErrInvalid, "invalid input"},
		{"ErrUnauthorized", ErrUnauthorized, "unauthorized"},
		{"ErrConflict", ErrConflict, "conflict"},
		{"ErrQuotaExceeded", ErrQuotaExceeded, "quota exceeded"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.value, "Error constant should not be nil")
			assert.Equal(t, tt.expected, tt.value.Error(), "Error constant should have expected value")
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "direct ErrNotFound",
			err:      ErrNotFound,
			expected: true,
		},
		{
			name:     "wrapped ErrNotFound",
			err:      fmt.Errorf("context: %w", ErrNotFound),
			expected: true,
		},
		{
			name:     "different error",
			err:      ErrInvalid,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "custom error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFound(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsInvalid(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "direct ErrInvalid",
			err:      ErrInvalid,
			expected: true,
		},
		{
			name:     "wrapped ErrInvalid",
			err:      fmt.Errorf("context: %w", ErrInvalid),
			expected: true,
		},
		{
			name:     "different error",
			err:      ErrNotFound,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "custom error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInvalid(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsUnauthorized(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "direct ErrUnauthorized",
			err:      ErrUnauthorized,
			expected: true,
		},
		{
			name:     "wrapped ErrUnauthorized",
			err:      fmt.Errorf("context: %w", ErrUnauthorized),
			expected: true,
		},
		{
			name:     "different error",
			err:      ErrInvalid,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "custom error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUnauthorized(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsConflict(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "direct ErrConflict",
			err:      ErrConflict,
			expected: true,
		},
		{
			name:     "wrapped ErrConflict",
			err:      fmt.Errorf("context: %w", ErrConflict),
			expected: true,
		},
		{
			name:     "different error",
			err:      ErrUnauthorized,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "custom error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConflict(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsQuotaExceeded(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "direct ErrQuotaExceeded",
			err:      ErrQuotaExceeded,
			expected: true,
		},
		{
			name:     "wrapped ErrQuotaExceeded",
			err:      fmt.Errorf("context: %w", ErrQuotaExceeded),
			expected: true,
		},
		{
			name:     "different error",
			err:      ErrInvalid,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "custom error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsQuotaExceeded(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsFunctions_Table(t *testing.T) {
	tests := []struct {
		name              string
		err               error
		expectNotFound    bool
		expectInvalid     bool
		expectUnauth      bool
		expectConflict    bool
		expectQuotaExceed bool
	}{
		{
			name:              "ErrNotFound",
			err:               ErrNotFound,
			expectNotFound:    true,
			expectInvalid:     false,
			expectUnauth:      false,
			expectConflict:    false,
			expectQuotaExceed: false,
		},
		{
			name:              "ErrInvalid",
			err:               ErrInvalid,
			expectNotFound:    false,
			expectInvalid:     true,
			expectUnauth:      false,
			expectConflict:    false,
			expectQuotaExceed: false,
		},
		{
			name:              "ErrUnauthorized",
			err:               ErrUnauthorized,
			expectNotFound:    false,
			expectInvalid:     false,
			expectUnauth:      true,
			expectConflict:    false,
			expectQuotaExceed: false,
		},
		{
			name:              "ErrConflict",
			err:               ErrConflict,
			expectNotFound:    false,
			expectInvalid:     false,
			expectUnauth:      false,
			expectConflict:    true,
			expectQuotaExceed: false,
		},
		{
			name:              "ErrQuotaExceeded",
			err:               ErrQuotaExceeded,
			expectNotFound:    false,
			expectInvalid:     false,
			expectUnauth:      false,
			expectConflict:    false,
			expectQuotaExceed: true,
		},
		{
			name:              "nil error",
			err:               nil,
			expectNotFound:    false,
			expectInvalid:     false,
			expectUnauth:      false,
			expectConflict:    false,
			expectQuotaExceed: false,
		},
		{
			name:              "custom error",
			err:               errors.New("custom error"),
			expectNotFound:    false,
			expectInvalid:     false,
			expectUnauth:      false,
			expectConflict:    false,
			expectQuotaExceed: false,
		},
		{
			name:              "wrapped ErrNotFound",
			err:               fmt.Errorf("wrapped: %w", ErrNotFound),
			expectNotFound:    true,
			expectInvalid:     false,
			expectUnauth:      false,
			expectConflict:    false,
			expectQuotaExceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectNotFound, IsNotFound(tt.err))
			assert.Equal(t, tt.expectInvalid, IsInvalid(tt.err))
			assert.Equal(t, tt.expectUnauth, IsUnauthorized(tt.err))
			assert.Equal(t, tt.expectConflict, IsConflict(tt.err))
			assert.Equal(t, tt.expectQuotaExceed, IsQuotaExceeded(tt.err))
		})
	}
}
