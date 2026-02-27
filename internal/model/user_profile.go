package model

// UserProfile represents user-specific configuration including Kindle email.
type UserProfile struct {
	Account     string `json:"account" dynamodbav:"account"`
	Email       string `json:"email" dynamodbav:"email"`
	KindleEmail string `json:"kindle_email" dynamodbav:"kindleEmail"`
}
