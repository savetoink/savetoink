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
	// schemeHTTP is the HTTP URL scheme.
	schemeHTTP = "http"
	// schemeHTTPS is the HTTPS URL scheme.
	schemeHTTPS = "https"
)

// ValidateURL validates URL format, scheme, host, length, and checks against private IPs.
func ValidateURL(urlStr string) error {
	if len(urlStr) > maxURLLength {
		return fmt.Errorf("%w: URL exceeds maximum length of %d characters", ErrInvalidURL, maxURLLength)
	}

	if urlStr == "" {
		return fmt.Errorf("%w: URL cannot be empty", ErrInvalidURL)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("%w: failed to parse URL", ErrInvalidURL)
	}

	if parsedURL.Scheme != schemeHTTP && parsedURL.Scheme != schemeHTTPS {
		return fmt.Errorf("%w: must use http or https scheme", ErrInvalidURL)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("%w: host is required", ErrInvalidURL)
	}

	if isPrivateHost(parsedURL.Host) {
		return fmt.Errorf("%w: %s", ErrPrivateIPAddress, parsedURL.Host)
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

// ValidateURLOnlyFormat performs basic URL format validation without private IP blocking.
// This is used in the extractor layer where URLs have already been validated at the handler level.
func ValidateURLOnlyFormat(urlStr string) error {
	if len(urlStr) > maxURLLength {
		return fmt.Errorf("%w: URL exceeds maximum length of %d characters", ErrInvalidURL, maxURLLength)
	}

	if urlStr == "" {
		return fmt.Errorf("%w: URL cannot be empty", ErrInvalidURL)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("%w: failed to parse URL", ErrInvalidURL)
	}

	if parsedURL.Scheme != schemeHTTP && parsedURL.Scheme != schemeHTTPS {
		return fmt.Errorf("%w: must use http or https scheme", ErrInvalidURL)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("%w: host is required", ErrInvalidURL)
	}

	return nil
}
