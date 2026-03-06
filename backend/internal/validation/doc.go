// Package validation provides input validation functions for URLs, emails, and device emails.
//
// It includes comprehensive validation for:
// - URL format, scheme, host, length, and private/internal IP blocking
// - Email format and length validation
// - Device email format with domain restrictions
//
// Exported error variables:
// - ErrInvalidURL: returned when URL validation fails
// - ErrInvalidEmail: returned when email validation fails
// - ErrPrivateIPAddress: returned when URL points to a private/internal network
//
// Exported constants:
// - MaxURLLength: maximum allowed length for URLs (2000 characters)
// - MaxEmailLength: maximum allowed length for email addresses (RFC 5321, 320 characters)
//
// Exported functions:
// - ValidateURL: validates URL format, scheme, host, and checks against private IPs
// - ValidateEmail: validates email format and length
// - ValidateDeviceEmail: validates email format and ensures it ends with a valid device email domain
package validation
