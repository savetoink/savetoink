package consts

// EmailProvider defines the email provider type.
type EmailProvider string

const (
	// EmailBackendMailjet indicates MailJet email backend.
	EmailBackendMailjet EmailProvider = "mailjet"
)

// Email constants.
const (
	// DefaultSubject is the default email subject.
	DefaultSubject = "Document"

	// MaxSubjectLength is the maximum length for email subjects.
	MaxSubjectLength = 100

	// DefaultBodyText is the default email body text.
	DefaultBodyText = "EPUB document attached."
)
