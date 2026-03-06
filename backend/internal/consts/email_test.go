package consts

import (
	"strings"
	"testing"
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
