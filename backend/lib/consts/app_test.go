package consts

import (
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// regexPattern is the regular expression pattern used to validate a semver string.
// Ref: https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
const regexPattern = `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`

func TestVersion(t *testing.T) {
	got := Version()

	require.NotNil(t, got, "Version() should return a non-nil pointer")
	assert.NotEmpty(t, *got, "Version() should return a non-empty string")
}

func TestVersion_MultipleCalls(t *testing.T) {
	v1 := Version()
	v2 := Version()

	assert.Equal(t, v1, v2, "Version() should return the same pointer on multiple calls")
}

func TestVersionFileIsVersionString(t *testing.T) {
	version := Version()
	require.NotNil(t, version, "Version() should not be nil")

	versionFile, err := os.ReadFile("../../../VERSION")
	require.NoError(t, err, "VERSION file should exist")
	versionFileStr := strings.TrimSuffix(string(versionFile), "\n")

	assert.Contains(t, *version, versionFileStr, "VERSION file should prefix the Version() string")
}

func TestVersionIsSemVer(t *testing.T) {
	version := Version()
	require.NotNil(t, version, "Version() should not be nil")

	re := regexp.MustCompile(regexPattern)
	matched := re.MatchString(*version)
	assert.True(t, matched, "Version() should return a valid semver string")
}

func TestRunMode_Constants(t *testing.T) {
	tests := []struct {
		name  string
		value RunMode
	}{
		{"ModeCLI", ModeCLI},
		{"ModeServer", ModeServer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, string(tt.value), "RunMode constant should not be empty")
		})
	}
}

func TestAuthBackend_Constants(t *testing.T) {
	tests := []struct {
		name  string
		value AuthBackend
	}{
		{"AuthBackendSharedAPIKey", AuthBackendSharedAPIKey},
		{"AuthBackendAuth0", AuthBackendAuth0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, string(tt.value), "AuthBackend constant should not be empty")
		})
	}
}

func TestLoggingProvider_Constants(t *testing.T) {
	tests := []struct {
		name  string
		value LoggingProvider
	}{
		{"LoggingBackendNone", LoggingBackendNone},
		{"LoggingBackendSentry", LoggingBackendSentry},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "LoggingBackendNone" {
				assert.Empty(t, string(tt.value), "LoggingBackendNone should be empty")
			} else {
				assert.NotEmpty(t, string(tt.value), "LoggingProvider constant should not be empty")
			}
		})
	}
}

func TestStatus_Constants(t *testing.T) {
	tests := []struct {
		name  string
		value Status
	}{
		{"StatusPending", StatusPending},
		{"StatusDelivered", StatusDelivered},
		{"StatusFailed", StatusFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, string(tt.value), "Status constant should not be empty")
		})
	}
}

func TestPagination_Constants(t *testing.T) {
	assert.GreaterOrEqual(t, DefaultPage, MinPage, "DefaultPage should be >= MinPage")
	assert.Equal(t, 1, MinPage, "MinPage should be 1")
	assert.GreaterOrEqual(t, DefaultPageSize, MinPageSize, "DefaultPageSize should be >= MinPageSize")
	assert.LessOrEqual(t, DefaultPageSize, MaxPageSize, "DefaultPageSize should be <= MaxPageSize")
	assert.Equal(t, 20, DefaultPageSize, "DefaultPageSize should be 20")
	assert.Equal(t, 20, MaxPageSize, "MaxPageSize should be 20")
}

func TestContentExtraction_Constants(t *testing.T) {
	assert.Greater(t, WordsPerMinute, 0, "WordsPerMinute should be positive")
	assert.Equal(t, 250, WordsPerMinute, "WordsPerMinute should be 250")
	assert.GreaterOrEqual(t, MinimumExtractedSize, 0, "MinimumExtractedSize should be non-negative")
	assert.GreaterOrEqual(t, MinimumOutputSize, 0, "MinimumOutputSize should be non-negative")
}

func TestAuth0ClientTimeout(t *testing.T) {
	assert.Greater(t, Auth0ClientTimeout, time.Duration(0), "Auth0ClientTimeout should be positive")
	assert.Equal(t, 10*time.Second, Auth0ClientTimeout, "Auth0ClientTimeout should be 10 seconds")
}

func TestFreeTier_Constants(t *testing.T) {
	assert.Greater(t, MaxFreeTierSendsPerPeriod, 0, "MaxFreeTierSendsPerPeriod should be positive")
	assert.Greater(t, FreeTierSendPeriodDays, 0, "FreeTierSendPeriodDays should be positive")
	assert.Equal(t, 10, MaxFreeTierSendsPerPeriod, "MaxFreeTierSendsPerPeriod should be 10")
	assert.Equal(t, 10, FreeTierSendPeriodDays, "FreeTierSendPeriodDays should be 10")
}
