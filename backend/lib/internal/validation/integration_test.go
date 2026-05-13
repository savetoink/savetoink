package validation

import (
	"testing"
)

const testNotAnEmail = "not-an-email"

func TestValidationIntegration(t *testing.T) {
	t.Run("URL validation prevents SSRF", func(t *testing.T) {
		badURLs := []string{
			"http://192.168.1.1/admin",
			"http://localhost:8080",
			"http://127.0.0.1/config",
			"http://10.0.0.1/secret",
			"http://169.254.169.254/metadata",
			"http://172.16.0.1/internal",
			"http://[::1]/loopback",
		}

		for _, urlStr := range badURLs {
			_, err := ValidateURL(urlStr)
			if err == nil {
				t.Errorf("Expected error for SSRF-protected URL: %s", urlStr)
			}
		}
	})

	t.Run("email validation prevents invalid emails", func(t *testing.T) {
		badEmails := []string{
			"",
			testNotAnEmail,
			"@example.com",
			"user@",
		}

		for _, email := range badEmails {
			err := ValidateEmail(email)
			if err == nil {
				t.Errorf("Expected error for invalid email: %s", email)
			}
		}
	})
}
