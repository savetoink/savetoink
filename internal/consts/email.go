package consts

// EmailProvider defines the email provider type.
type EmailProvider string

const (
	// EmailBackendMailjet indicates MailJet email backend.
	EmailBackendMailjet EmailProvider = "mailjet"
)

// Email constants.
const (
	// MaxSubjectLength is the maximum length for email subjects.
	MaxSubjectLength = 100

	// MailSubjectPrefix is the prefix for email subjects.
	MailSubjectPrefix = "[Save to Ink] "
)
