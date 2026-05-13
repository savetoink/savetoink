package apperrors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testDifferentError   = "different error"
	testSomeOtherError   = "some other error"
	testCustomError      = "custom error"
	testNilError         = "nil error"
	testWrappedNotFound  = "wrapped ErrNotFound"
	testErrNotFound      = "ErrNotFound"
	testErrInvalid       = "ErrInvalid"
	testErrUnauthorized  = "ErrUnauthorized"
	testErrConflict      = "ErrConflict"
	testErrQuotaExceeded = "ErrQuotaExceeded"
	wrappedErrorFmt      = "context: %w"
)

func TestError_Constants(t *testing.T) {
	tests := []struct {
		name     string
		value    error
		expected string
	}{
		{testErrNotFound, ErrNotFound, "not found"},
		{testErrInvalid, ErrInvalid, "invalid input"},
		{testErrUnauthorized, ErrUnauthorized, "unauthorized"},
		{testErrConflict, ErrConflict, "conflict"},
		{testErrQuotaExceeded, ErrQuotaExceeded, "quota exceeded"},
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
			name:     testWrappedNotFound,
			err:      fmt.Errorf(wrappedErrorFmt, ErrNotFound),
			expected: true,
		},
		{
			name:     testDifferentError,
			err:      ErrInvalid,
			expected: false,
		},
		{
			name:     testNilError,
			err:      nil,
			expected: false,
		},
		{
			name:     testCustomError,
			err:      errors.New(testSomeOtherError),
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
			err:      fmt.Errorf(wrappedErrorFmt, ErrInvalid),
			expected: true,
		},
		{
			name:     testDifferentError,
			err:      ErrNotFound,
			expected: false,
		},
		{
			name:     testNilError,
			err:      nil,
			expected: false,
		},
		{
			name:     testCustomError,
			err:      errors.New(testSomeOtherError),
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
			err:      fmt.Errorf(wrappedErrorFmt, ErrUnauthorized),
			expected: true,
		},
		{
			name:     testDifferentError,
			err:      ErrInvalid,
			expected: false,
		},
		{
			name:     testNilError,
			err:      nil,
			expected: false,
		},
		{
			name:     testCustomError,
			err:      errors.New(testSomeOtherError),
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
			err:      fmt.Errorf(wrappedErrorFmt, ErrConflict),
			expected: true,
		},
		{
			name:     testDifferentError,
			err:      ErrUnauthorized,
			expected: false,
		},
		{
			name:     testNilError,
			err:      nil,
			expected: false,
		},
		{
			name:     testCustomError,
			err:      errors.New(testSomeOtherError),
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
			err:      fmt.Errorf(wrappedErrorFmt, ErrQuotaExceeded),
			expected: true,
		},
		{
			name:     testDifferentError,
			err:      ErrInvalid,
			expected: false,
		},
		{
			name:     testNilError,
			err:      nil,
			expected: false,
		},
		{
			name:     testCustomError,
			err:      errors.New(testSomeOtherError),
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
			name:              testErrNotFound,
			err:               ErrNotFound,
			expectNotFound:    true,
			expectInvalid:     false,
			expectUnauth:      false,
			expectConflict:    false,
			expectQuotaExceed: false,
		},
		{
			name:              testErrInvalid,
			err:               ErrInvalid,
			expectNotFound:    false,
			expectInvalid:     true,
			expectUnauth:      false,
			expectConflict:    false,
			expectQuotaExceed: false,
		},
		{
			name:              testErrUnauthorized,
			err:               ErrUnauthorized,
			expectNotFound:    false,
			expectInvalid:     false,
			expectUnauth:      true,
			expectConflict:    false,
			expectQuotaExceed: false,
		},
		{
			name:              testErrConflict,
			err:               ErrConflict,
			expectNotFound:    false,
			expectInvalid:     false,
			expectUnauth:      false,
			expectConflict:    true,
			expectQuotaExceed: false,
		},
		{
			name:              testErrQuotaExceeded,
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
			name:              testCustomError,
			err:               errors.New(testCustomError),
			expectNotFound:    false,
			expectInvalid:     false,
			expectUnauth:      false,
			expectConflict:    false,
			expectQuotaExceed: false,
		},
		{
			name:              testWrappedNotFound,
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
