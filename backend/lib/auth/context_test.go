package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAccountID(t *testing.T) {
	t.Run("returns account ID when set", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), AccountIDKey, "test-account-123")
		result := GetAccountID(ctx)
		assert.Equal(t, "test-account-123", result)
	})

	t.Run("returns empty string when not set", func(t *testing.T) {
		ctx := context.Background()
		result := GetAccountID(ctx)
		assert.Equal(t, "", result)
	})

	t.Run("returns empty string when wrong type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), AccountIDKey, 123)
		result := GetAccountID(ctx)
		assert.Equal(t, "", result)
	})
}

func TestGetAuthError(t *testing.T) {
	t.Run("returns error when set", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), AuthErrorKey, "authentication failed")
		result := GetAuthError(ctx)
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "authentication failed")
	})

	t.Run("returns nil when not set", func(t *testing.T) {
		ctx := context.Background()
		result := GetAuthError(ctx)
		assert.NoError(t, result)
	})
}

func TestGetSendsCount(t *testing.T) {
	t.Run("returns count when set", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), SendsCountKey, 5)
		result := GetSendsCount(ctx)
		assert.Equal(t, 5, result)
	})

	t.Run("returns 0 when not set", func(t *testing.T) {
		ctx := context.Background()
		result := GetSendsCount(ctx)
		assert.Equal(t, 0, result)
	})
}

func TestHasSendsCount(t *testing.T) {
	t.Run("returns true when set", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), SendsCountKey, 5)
		result := HasSendsCount(ctx)
		assert.True(t, result)
	})

	t.Run("returns false when not set", func(t *testing.T) {
		ctx := context.Background()
		result := HasSendsCount(ctx)
		assert.False(t, result)
	})
}
