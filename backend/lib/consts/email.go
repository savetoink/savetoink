package consts

import (
	"fmt"
	"strings"
)

// EmailProvider defines the email provider type.
type EmailProvider string

const (
	// EmailBackendMailjet indicates MailJet email backend.
	EmailBackendMailjet EmailProvider = "mailjet"
)

// validDeviceEmailDomains contains the valid email domain suffixes for device emails.
var validDeviceEmailDomains = []string{
	"@kindle.com",
	"@free.kindle.com",
	"@send.kobo.com",
	"@pbsync.com",
	"@mytolino.com",
}

// GetValidDeviceEmailDomains returns a copy of the valid device email domains slice.
func GetValidDeviceEmailDomains() []string {
	domains := make([]string, len(validDeviceEmailDomains))
	copy(domains, validDeviceEmailDomains)
	return domains
}

// ValidDeviceEmailDomainsJoined returns the valid domains joined with " or ".
func ValidDeviceEmailDomainsJoined() string {
	return strings.Join(validDeviceEmailDomains, " or ")
}

// Email constants.
const (
	// MaxSubjectLength is the maximum length for email subjects.
	MaxSubjectLength = 100

	// MailSubjectPrefix is the prefix for email subjects.
	MailSubjectPrefix   = "[Save to Ink] "
	DefaultEmailSubject = MailSubjectPrefix + "Article"

	// LandingURL is the URL for the landing page.
	LandingURL = "https://www.saveto.ink"
)

// BuildEmailBody creates the email body text with the app URL and landing URL.
func BuildEmailBody(appURL string) string {
	return fmt.Sprintf(`EPUB document attached.

To disable email delivery update your account settings at %s

---
Save to Ink - %s
 `, appURL, LandingURL)
}

// BuildCLIEmailBody creates a simpler email body for CLI mode.
func BuildCLIEmailBody() string {
	return fmt.Sprintf(`EPUB document attached.

---
Save to Ink - %s
 `, LandingURL)
}
