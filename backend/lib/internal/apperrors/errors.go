// Package apperrors provides sentinel errors and error handling conventions for the application.
package apperrors

import "errors"

// Sentinel errors for common scenarios.
var (
	// ErrNotFound indicates a requested resource was not found.
	ErrNotFound = errors.New("not found")

	// ErrInvalid indicates invalid input or validation failure.
	ErrInvalid = errors.New("invalid input")

	// ErrUnauthorized indicates authentication or authorization failure.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrConflict indicates a state conflict (e.g., duplicate creation).
	ErrConflict = errors.New("conflict")

	// ErrQuotaExceeded indicates a quota limit has been exceeded.
	ErrQuotaExceeded = errors.New("quota exceeded")
)

// IsNotFound checks if error is ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsInvalid checks if error is ErrInvalid.
func IsInvalid(err error) bool {
	return errors.Is(err, ErrInvalid)
}

// IsUnauthorized checks if error is ErrUnauthorized.
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsConflict checks if error is ErrConflict.
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsQuotaExceeded checks if error is ErrQuotaExceeded.
func IsQuotaExceeded(err error) bool {
	return errors.Is(err, ErrQuotaExceeded)
}
