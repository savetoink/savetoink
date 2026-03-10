package validation

import (
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"strings"

	"github.com/shaftoe/savetoink/backend/lib/consts"
)

var (
	// ErrInvalidURL is returned when URL validation fails.
	ErrInvalidURL = errors.New("invalid URL")
	// ErrInvalidEmail is returned when email validation fails.
	ErrInvalidEmail = errors.New("invalid email address")
	// ErrPrivateIPAddress is returned when URL points to a private/internal network.
	ErrPrivateIPAddress = errors.New("URL points to private/internal network")
)

const (
	// maxURLLength is the maximum allowed length for URLs.
	maxURLLength = 2000
	// maxEmailLength is the maximum allowed length for email addresses (RFC 5321).
	maxEmailLength = 320
	// hostSplitMaxParts is the maximum number of parts to split when extracting hostname from host:port.
	hostSplitMaxParts = 2
	// schemeHTTP is HTTP URL scheme.
	schemeHTTP = "http"
	// schemeHTTPS is HTTPS URL scheme.
	schemeHTTPS = "https"
)

// ValidateURL parses and validates a URL string, returning a validated *url.URL.
// This performs basic validation (scheme, host) and is used at entry points.
func ValidateURL(rawURL string) (*url.URL, error) {
	if len(rawURL) > maxURLLength {
		return nil, fmt.Errorf("%w: URL exceeds maximum length of %d characters", ErrInvalidURL, maxURLLength)
	}

	if rawURL == "" {
		return nil, fmt.Errorf("%w: URL cannot be empty", ErrInvalidURL)
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse URL", ErrInvalidURL)
	}

	if parsedURL.Scheme != schemeHTTP && parsedURL.Scheme != schemeHTTPS {
		return nil, fmt.Errorf("%w: must use http or https scheme", ErrInvalidURL)
	}

	if parsedURL.Host == "" {
		return nil, fmt.Errorf("%w: host is required", ErrInvalidURL)
	}

	if isPrivateHost(parsedURL.Host) {
		return nil, fmt.Errorf("%w: %s", ErrPrivateIPAddress, parsedURL.Host)
	}

	return parsedURL, nil
}

// ValidateParsedURL performs business validation on an already-parsed URL.
// This checks length and private IP addresses, and is used internally
// after URL has been parsed by ValidateURL().
func ValidateParsedURL(u *url.URL) error {
	if len(u.String()) > maxURLLength {
		return fmt.Errorf("%w: URL exceeds maximum length of %d characters", ErrInvalidURL, maxURLLength)
	}

	if isPrivateHost(u.Host) {
		return fmt.Errorf("%w: %s", ErrPrivateIPAddress, u.Host)
	}

	return nil
}

// isPrivateHost checks if a host points to a private/internal network.
func isPrivateHost(host string) bool {
	host = strings.TrimSuffix(host, ":80")
	host = strings.TrimSuffix(host, ":443")

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	} else if strings.Contains(host, ":") {
		parts := strings.SplitN(host, ":", hostSplitMaxParts)
		host = parts[0]
	}

	if strings.EqualFold(host, "localhost") {
		return true
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast()
}

// ValidateEmail validates email format and length.
func ValidateEmail(email string) error {
	if len(email) > maxEmailLength {
		return fmt.Errorf("%w: email exceeds maximum length of %d characters", ErrInvalidEmail, maxEmailLength)
	}

	if email == "" {
		return fmt.Errorf("%w: email cannot be empty", ErrInvalidEmail)
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("%w: invalid email format", ErrInvalidEmail)
	}

	if addr.Address != email {
		return fmt.Errorf("%w: invalid email format", ErrInvalidEmail)
	}

	return nil
}

// ValidateDeviceEmail validates email format and ensures it ends with a valid device email domain.
func ValidateDeviceEmail(email string) error {
	if err := ValidateEmail(email); err != nil {
		return err
	}

	domains := consts.GetValidDeviceEmailDomains()
	for _, domain := range domains {
		if strings.HasSuffix(email, domain) {
			return nil
		}
	}

	return fmt.Errorf(
		"%w: must be a valid email ending with %s",
		ErrInvalidEmail,
		consts.ValidDeviceEmailDomainsJoined(),
	)
}
