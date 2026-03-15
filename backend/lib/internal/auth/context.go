// Package auth provides authentication context helpers for the savetoink application.
package auth

import (
	"context"
	"errors"
)

type contextKey string

const (
	accountIDKey contextKey = "account_id"
	authErrorKey contextKey = "auth_error"
)

// GetAccountIDFromCtx retrieves authenticated account ID from context.
func GetAccountIDFromCtx(ctx context.Context) string {
	accountID, _ := ctx.Value(accountIDKey).(string)
	return accountID
}

// GetAuthErrorFromCtx retrieves authentication error from context, if any.
func GetAuthErrorFromCtx(ctx context.Context) error {
	authError, found := ctx.Value(authErrorKey).(string)
	if found {
		return errors.New(authError)
	}
	return nil
}

// AddAccountIDToCtx adds an account ID to the context.
func AddAccountIDToCtx(ctx context.Context, accountID string) context.Context {
	return context.WithValue(ctx, accountIDKey, accountID)
}

// AddAuthErrorToCtx adds an authentication error to the context.
func AddAuthErrorToCtx(ctx context.Context, msg string) context.Context {
	return context.WithValue(ctx, authErrorKey, msg)
}
