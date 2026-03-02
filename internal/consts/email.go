package consts

import "strings"

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
	MailSubjectPrefix = "[Save to Ink] "

	// LandingURL is the URL for the landing page.
	LandingURL = "https://www.saveto.ink"
)
