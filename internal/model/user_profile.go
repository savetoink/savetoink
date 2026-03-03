package model

import "time"

// BounceInfo stores information about a bounced email.
type BounceInfo struct {
	Timestamp time.Time `json:"timestamp" dynamodbav:"timestamp"`
	Error     string    `json:"error" dynamodbav:"error"`
}

// UserProfile represents user-specific configuration.
type UserProfile struct {
	Account       string                `json:"account" dynamodbav:"account"`
	Email         string                `json:"email" dynamodbav:"email"`
	DeviceEmail   string                `json:"device_email" dynamodbav:"deviceEmail"`
	AutoSend      bool                  `json:"auto_send" dynamodbav:"autoSend"`
	BouncedEmails map[string]BounceInfo `json:"bounced_emails" dynamodbav:"bouncedEmails,omitempty"`
}
