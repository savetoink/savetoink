package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAccountID(t *testing.T) {
	t.Run("returns account ID when set", func(t *testing.T) {
		ctx := AddAccountIDToCtx(context.Background(), "test-account-123")
		result := GetAccountIDFromCtx(ctx)
		assert.Equal(t, "test-account-123", result)
	})

	t.Run("returns empty string when not set", func(t *testing.T) {
		ctx := context.Background()
		result := GetAccountIDFromCtx(ctx)
		assert.Equal(t, "", result)
	})
}

func TestGetAuthError(t *testing.T) {
	t.Run("returns error when set", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), authErrorKey, "authentication failed")
		result := GetAuthErrorFromCtx(ctx)
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "authentication failed")
	})

	t.Run("returns nil when not set", func(t *testing.T) {
		ctx := context.Background()
		result := GetAuthErrorFromCtx(ctx)
		assert.NoError(t, result)
	})
}
