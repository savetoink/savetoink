package consts

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testAppURL = "https://app.saveto.ink"

func TestBuildEmailBody(t *testing.T) {
	tests := []struct {
		name   string
		appURL string
	}{
		{
			name:   "valid app url",
			appURL: testAppURL,
		},
		{
			name:   "empty app url",
			appURL: "",
		},
		{
			name:   "long app url",
			appURL: "https://very-long-app-url.example.com/with/path?query=params",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildEmailBody(tt.appURL)
			if got == "" {
				t.Error("BuildEmailBody() returned empty string")
			}
		})
	}
}

func TestBuildEmailBody_ContainsLandingURL(t *testing.T) {
	got := BuildEmailBody(testAppURL)

	expectedLanding := LandingURL
	if !strings.Contains(got, expectedLanding) {
		t.Errorf("BuildEmailBody() result does not contain landing URL %s", expectedLanding)
	}
}

func TestBuildEmailBody_ContainsAppURL(t *testing.T) {
	got := BuildEmailBody(testAppURL)

	if !strings.Contains(got, testAppURL) {
		t.Errorf("BuildEmailBody() result does not contain app URL %s", testAppURL)
	}
}

func TestBuildCLIEmailBody(t *testing.T) {
	got := BuildCLIEmailBody()

	if got == "" {
		t.Error("BuildCLIEmailBody() returned empty string")
	}
}

func TestBuildCLIEmailBody_ContainsLandingURL(t *testing.T) {
	got := BuildCLIEmailBody()

	if !strings.Contains(got, LandingURL) {
		t.Errorf("BuildCLIEmailBody() result does not contain landing URL %s", LandingURL)
	}
}

func TestBuildCLIEmailBody_NoAppURL(t *testing.T) {
	got := BuildCLIEmailBody()

	if strings.Contains(got, testAppURL) {
		t.Errorf("BuildCLIEmailBody() result should not contain app URL %s", testAppURL)
	}

	if strings.Contains(got, "account settings") {
		t.Error("BuildCLIEmailBody() result should not contain account settings text")
	}
}

func TestGetValidDeviceEmailDomains(t *testing.T) {
	domains := GetValidDeviceEmailDomains()

	if len(domains) == 0 {
		t.Error("GetValidDeviceEmailDomains() returned empty slice")
	}

	for _, domain := range domains {
		if domain == "" {
			t.Error("GetValidDeviceEmailDomains() returned empty domain")
		}
		if !strings.HasPrefix(domain, "@") {
			t.Errorf("GetValidDeviceEmailDomains() returned domain without @ prefix: %s", domain)
		}
	}
}

func TestGetValidDeviceEmailDomains_ReturnsCopy(t *testing.T) {
	domains1 := GetValidDeviceEmailDomains()
	domains2 := GetValidDeviceEmailDomains()

	if &domains1[0] == &domains2[0] {
		t.Error("GetValidDeviceEmailDomains() should return a copy, not the same slice")
	}
}

func TestValidDeviceEmailDomainsJoined(t *testing.T) {
	joined := ValidDeviceEmailDomainsJoined()

	if joined == "" {
		t.Error("ValidDeviceEmailDomainsJoined() returned empty string")
	}

	if !strings.Contains(joined, "@kindle.com") {
		t.Error("ValidDeviceEmailDomainsJoined() should contain @kindle.com")
	}

	if !strings.Contains(joined, " or ") {
		t.Error("ValidDeviceEmailDomainsJoined() should contain ' or ' separator")
	}
}

func TestEmail_Constants(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"EmailBackendMailjet", string(EmailBackendMailjet), "mailjet"},
		{"MailSubjectPrefix", MailSubjectPrefix, "[Save to Ink] "},
		{"LandingURL", LandingURL, "https://www.saveto.ink"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.value, "Email constant should not be empty")
			assert.Equal(t, tt.expected, tt.value, "Email constant should have expected value")
		})
	}
}

func TestMaxSubjectLength(t *testing.T) {
	assert.Greater(t, MaxSubjectLength, 0, "MaxSubjectLength should be positive")
	assert.Equal(t, 100, MaxSubjectLength, "MaxSubjectLength should be 100")
}
