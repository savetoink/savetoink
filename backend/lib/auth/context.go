// Package auth provides authentication context helpers for the savetoink application.
package auth

import (
	"context"
	"errors"
)

type contextKey string

const (
	// AccountIDKey is the context key for account ID.
	AccountIDKey contextKey = "account_id"
	// AuthErrorKey is the context key for authentication error.
	AuthErrorKey contextKey = "auth_error"
)

// GetAccountID retrieves authenticated account ID from context.
func GetAccountID(ctx context.Context) string {
	accountID, _ := ctx.Value(AccountIDKey).(string)
	return accountID
}

// GetAuthError retrieves authentication error from context, if any.
func GetAuthError(ctx context.Context) error {
	authError, found := ctx.Value(AuthErrorKey).(string)
	if found {
		return errors.New(authError)
	}
	return nil
}
