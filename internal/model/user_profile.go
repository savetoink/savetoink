package model

// UserProfile represents user-specific configuration.
type UserProfile struct {
	Account     string `json:"account" dynamodbav:"account"`
	Email       string `json:"email" dynamodbav:"email"`
	DeviceEmail string `json:"device_email" dynamodbav:"deviceEmail"`
	AutoSend    bool   `json:"auto_send" dynamodbav:"autoSend"`
}
