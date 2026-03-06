package config

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/shaftoe/savetoink/backend/internal/consts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	_ = os.Unsetenv("SAVETOINK_API_KEY")
	_ = os.Unsetenv("SAVETOINK_ARTICLE_TABLE_NAME")
	_ = os.Unsetenv("SAVETOINK_USER_PROFILE_TABLE_NAME")
	_ = os.Unsetenv("SAVETOINK_SENDS_TABLE_NAME")
	_ = os.Unsetenv("SAVETOINK_APP_URL")
	_ = os.Unsetenv("SAVETOINK_EMAIL_BACKEND")
	_ = os.Unsetenv("SAVETOINK_MAILJET_API_KEY")
	_ = os.Unsetenv("SAVETOINK_MAILJET_API_SECRET")
	_ = os.Unsetenv("SAVETOINK_MAILJET_WEBHOOK_SECRET")
	_ = os.Unsetenv("SAVETOINK_SENDER_EMAIL")
	_ = os.Unsetenv("SAVETOINK_AUTH_BACKEND")
	_ = os.Unsetenv("SAVETOINK_AUTH0_DOMAIN")
	_ = os.Unsetenv("SAVETOINK_AUTH0_AUDIENCE")
	_ = os.Unsetenv("SAVETOINK_AUTH0_CLIENT_ID")
	_ = os.Unsetenv("SAVETOINK_AUTH0_CLIENT_SECRET")
	_ = os.Unsetenv("SAVETOINK_LOGGING_PROVIDER")
	_ = os.Unsetenv("SAVETOINK_SENTRY_DSN")
	_ = os.Unsetenv("SAVETOINK_SENTRY_ENVIRONMENT")
	_ = os.Unsetenv("SAVETOINK_SENTRY_SAMPLE_RATE")
	_ = os.Unsetenv("SAVETOINK_DEBUG")
	os.Exit(m.Run())
}

func TestLoad_CLI_Mode(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_DEBUG": "false",
	})

	cfg, err := Load(consts.ModeCLI, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.ModeCLI, cfg.Mode)
	assert.False(t, cfg.Debug)
	assert.Nil(t, cfg.AWSConfig)
}

func TestLoad_CLI_Mode_With_Debug(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_DEBUG": "true",
	})

	cfg, err := Load(consts.ModeCLI, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.ModeCLI, cfg.Mode)
	assert.True(t, cfg.Debug)
}

func TestLoad_Server_Mode_Success(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
		"SAVETOINK_EMAIL_BACKEND":           "mailjet",
		"SAVETOINK_MAILJET_API_KEY":         "mailjet-key",
		"SAVETOINK_MAILJET_API_SECRET":      "mailjet-secret",
		"SAVETOINK_MAILJET_WEBHOOK_SECRET":  "webhook-secret",
		"SAVETOINK_SENDER_EMAIL":            "sender@example.com",
	})

	awsLoader := mockAWSLoader()

	cfg, err := Load(consts.ModeServer, awsLoader)
	require.NoError(t, err)
	assert.Equal(t, consts.ModeServer, cfg.Mode)
	assert.NotNil(t, cfg.AWSConfig)
	assert.Equal(t, "test-api-key", cfg.APIKeySecret)
	assert.Equal(t, "articles-table", cfg.ArticlesTable)
	assert.Equal(t, "profiles-table", cfg.UserProfileTable)
	assert.Equal(t, "sends-table", cfg.SendsTable)
	assert.Equal(t, "https://example.com", cfg.AppURL)
	assert.Equal(t, consts.EmailBackendMailjet, cfg.EmailProvider)
}

func TestLoad_Default_Auth_Backend(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
	})

	awsLoader := mockAWSLoader()

	cfg, err := Load(consts.ModeServer, awsLoader)
	require.NoError(t, err)
	assert.Equal(t, consts.AuthBackendSharedAPIKey, cfg.AuthBackend)
}

func TestLoad_Auth0_Backend_Success(t *testing.T) {
	setupEnvVars(t, map[string]string{ //nolint:gosec // Test values, not real credentials
		"SAVETOINK_AUTH_BACKEND":            "auth0",
		"SAVETOINK_AUTH0_DOMAIN":            "auth0-domain",
		"SAVETOINK_AUTH0_AUDIENCE":          "auth0-audience",
		"SAVETOINK_AUTH0_CLIENT_ID":         "auth0-client-id",
		"SAVETOINK_AUTH0_CLIENT_SECRET":     "auth0-client-secret",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
	})

	awsLoader := mockAWSLoader()

	cfg, err := Load(consts.ModeServer, awsLoader)
	require.NoError(t, err)
	assert.Equal(t, consts.AuthBackendAuth0, cfg.AuthBackend)
	assert.Equal(t, "auth0-domain", cfg.Auth0Domain)
	assert.Equal(t, "auth0-audience", cfg.Auth0Audience)
	assert.Equal(t, "auth0-client-id", cfg.Auth0ClientID)
	assert.Equal(t, "auth0-client-secret", cfg.Auth0ClientSecret)
}

func TestLoad_Invalid_Auth_Backend(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_AUTH_BACKEND":       "invalid-backend",
		"SAVETOINK_ARTICLE_TABLE_NAME": "articles-table",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported auth backend")
}

func TestLoad_Missing_Required_Server_Env(t *testing.T) {
	setupEnvVars(t, map[string]string{})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required environment variables are missing")
}

func TestLoad_Missing_API_Key_For_Shared_API_Key(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_ARTICLE_TABLE_NAME": "articles-table",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SAVETOINK_API_KEY")
}

func TestLoad_Missing_Multiple_Auth0_Env(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_AUTH_BACKEND":       "auth0",
		"SAVETOINK_ARTICLE_TABLE_NAME": "articles-table",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	errMsg := err.Error()
	assert.Contains(t, errMsg, "SAVETOINK_AUTH0_DOMAIN")
	assert.Contains(t, errMsg, "SAVETOINK_AUTH0_AUDIENCE")
	assert.Contains(t, errMsg, "SAVETOINK_AUTH0_CLIENT_ID")
	assert.Contains(t, errMsg, "SAVETOINK_AUTH0_CLIENT_SECRET")
}

func TestLoad_Missing_ArticlesTable(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY": "test-api-key",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SAVETOINK_ARTICLE_TABLE_NAME")
}

func TestLoad_Missing_UserProfileTable(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY":            "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME": "articles-table",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SAVETOINK_USER_PROFILE_TABLE_NAME")
}

func TestLoad_Missing_SendsTable(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY":            "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME": "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE": "profiles-table",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SAVETOINK_SENDS_TABLE_NAME")
}

func TestLoad_Missing_AppURL(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY":            "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME": "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":   "sends-table",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SAVETOINK_APP_URL")
}

func TestLoad_Missing_Mailjet_API_Key(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY":            "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME": "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":   "sends-table",
		"SAVETOINK_APP_URL":            "https://example.com",
		"SAVETOINK_EMAIL_BACKEND":      "mailjet",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SAVETOINK_MAILJET_API_KEY")
}

func TestLoad_Missing_Mailjet_API_Secret(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY":            "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME": "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":   "sends-table",
		"SAVETOINK_APP_URL":            "https://example.com",
		"SAVETOINK_EMAIL_BACKEND":      "mailjet",
		"SAVETOINK_MAILJET_API_KEY":    "mailjet-key",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SAVETOINK_MAILJET_API_SECRET")
}

func TestLoad_Missing_Mailjet_Webhook_Secret(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY":            "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME": "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":   "sends-table",
		"SAVETOINK_APP_URL":            "https://example.com",
		"SAVETOINK_EMAIL_BACKEND":      "mailjet",
		"SAVETOINK_MAILJET_API_KEY":    "mailjet-key",
		"SAVETOINK_MAILJET_API_SECRET": "mailjet-secret",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SAVETOINK_MAILJET_WEBHOOK_SECRET")
}

func TestLoad_Missing_Sender_Email(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY":             "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":  "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE":  "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":    "sends-table",
		"SAVETOINK_APP_URL":             "https://example.com",
		"SAVETOINK_EMAIL_BACKEND":       "mailjet",
		"SAVETOINK_MAILJET_API_KEY":     "mailjet-key",
		"SAVETOINK_MAILJET_API_SECRET":  "mailjet-secret",
		"SAVETOINK_MAILJET_WEBHOOK_SEC": "webhook-secret",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SAVETOINK_SENDER_EMAIL")
}

func TestLoad_Sentry_Provider_Success(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
		"SAVETOINK_LOGGING_PROVIDER":        "sentry",
		"SAVETOINK_SENTRY_DSN":              "sentry-dsn",
		"SAVETOINK_SENTRY_ENVIRONMENT":      "production",
		"SAVETOINK_SENTRY_SAMPLE_RATE":      "1.0",
	})

	awsLoader := mockAWSLoader()

	cfg, err := Load(consts.ModeServer, awsLoader)
	require.NoError(t, err)
	assert.Equal(t, consts.LoggingBackendSentry, cfg.LoggingProvider)
	assert.Equal(t, "sentry-dsn", cfg.SentryDSN)
	assert.Equal(t, "production", cfg.SentryEnvironment)
	assert.Equal(t, 1.0, cfg.SentrySampleRate)
}

func TestLoad_Missing_Sentry_DSN(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY":            "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME": "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":   "sends-table",
		"SAVETOINK_APP_URL":            "https://example.com",
		"SAVETOINK_LOGGING_PROVIDER":   "sentry",
		"SAVETOINK_SENTRY_ENVIRONMENT": "production",
		"SAVETOINK_SENTRY_SAMPLE_RATE": "1.0",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SAVETOINK_SENTRY_DSN")
}

func TestLoad_Missing_Sentry_Environment(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY":            "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME": "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":   "sends-table",
		"SAVETOINK_APP_URL":            "https://example.com",
		"SAVETOINK_LOGGING_PROVIDER":   "sentry",
		"SAVETOINK_SENTRY_DSN":         "sentry-dsn",
		"SAVETOINK_SENTRY_SAMPLE_RATE": "1.0",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SAVETOINK_SENTRY_ENVIRONMENT")
}

func TestLoad_Missing_Sentry_Sample_Rate(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY":            "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME": "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":   "sends-table",
		"SAVETOINK_APP_URL":            "https://example.com",
		"SAVETOINK_LOGGING_PROVIDER":   "sentry",
		"SAVETOINK_SENTRY_DSN":         "sentry-dsn",
		"SAVETOINK_SENTRY_ENVIRONMENT": "production",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SAVETOINK_SENTRY_SAMPLE_RATE")
}

func TestLoad_Non_Sentry_Logging_Does_Not_Validate_Sentry(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
	})

	awsLoader := mockAWSLoader()

	cfg, err := Load(consts.ModeServer, awsLoader)
	require.NoError(t, err)
	assert.Equal(t, consts.LoggingBackendNone, cfg.LoggingProvider)
	assert.Empty(t, cfg.SentryDSN)
	assert.Empty(t, cfg.SentryEnvironment)
	assert.Equal(t, 0.0, cfg.SentrySampleRate)
}

func TestLoad_Multiple_Validation_Errors(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_AUTH_BACKEND": "auth0",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	errMsg := err.Error()
	assert.Contains(t, errMsg, "SAVETOINK_AUTH0_DOMAIN")
	assert.Contains(t, errMsg, "SAVETOINK_AUTH0_AUDIENCE")
	assert.Contains(t, errMsg, "SAVETOINK_AUTH0_CLIENT_ID")
	assert.Contains(t, errMsg, "SAVETOINK_AUTH0_CLIENT_SECRET")
	assert.Contains(t, errMsg, "SAVETOINK_ARTICLE_TABLE_NAME")
	assert.Contains(t, errMsg, "SAVETOINK_USER_PROFILE_TABLE_NAME")
	assert.Contains(t, errMsg, "SAVETOINK_SENDS_TABLE_NAME")
	assert.Contains(t, errMsg, "SAVETOINK_APP_URL")
}

func TestLoad_Server_Mode_Without_AWS_Loader(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY":            "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME": "articles-table",
	})

	_, err := Load(consts.ModeServer, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AWS config loader is required")
}

func TestLoad_Server_Mode_AWS_Loader_Error(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY":            "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME": "articles-table",
	})

	awsLoader := func(_ context.Context) (aws.Config, error) {
		return aws.Config{}, assert.AnError
	}

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load AWS config")
}

func TestLoad_All_Env_Vars_Bound(t *testing.T) {
	envVars := map[string]string{ //nolint:gosec // Test values, not real credentials
		"SAVETOINK_API_KEY":                 "api-key-secret",
		"SAVETOINK_APP_URL":                 "https://example.com",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_AUTH0_AUDIENCE":          "auth0-audience",
		"SAVETOINK_AUTH0_CLIENT_ID":         "auth0-client-id",
		"SAVETOINK_AUTH0_CLIENT_SECRET":     "auth0-client-secret",
		"SAVETOINK_AUTH0_DOMAIN":            "auth0-domain",
		"SAVETOINK_AUTH_BACKEND":            "shared_api_key",
		"SAVETOINK_DEBUG":                   "true",
		"SAVETOINK_EMAIL_BACKEND":           "mailjet",
		"SAVETOINK_MAILJET_API_KEY":         "mailjet-key",
		"SAVETOINK_MAILJET_API_SECRET":      "mailjet-secret",
		"SAVETOINK_MAILJET_WEBHOOK_SECRET":  "webhook-secret",
		"SAVETOINK_SENDER_EMAIL":            "sender@example.com",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_LOGGING_PROVIDER":        "sentry",
		"SAVETOINK_SENTRY_DSN":              "sentry-dsn",
		"SAVETOINK_SENTRY_ENVIRONMENT":      "production",
		"SAVETOINK_SENTRY_SAMPLE_RATE":      "0.5",
	}

	setupEnvVars(t, envVars)

	awsLoader := mockAWSLoader()

	cfg, err := Load(consts.ModeServer, awsLoader)
	require.NoError(t, err)

	assert.Equal(t, "api-key-secret", cfg.APIKeySecret)
	assert.Equal(t, "https://example.com", cfg.AppURL)
	assert.Equal(t, "articles-table", cfg.ArticlesTable)
	assert.Equal(t, "auth0-audience", cfg.Auth0Audience)
	assert.Equal(t, "auth0-client-id", cfg.Auth0ClientID)
	assert.Equal(t, "auth0-client-secret", cfg.Auth0ClientSecret)
	assert.Equal(t, "auth0-domain", cfg.Auth0Domain)
	assert.Equal(t, consts.AuthBackendSharedAPIKey, cfg.AuthBackend)
	assert.True(t, cfg.Debug)
	assert.Equal(t, consts.EmailBackendMailjet, cfg.EmailProvider)
	assert.Equal(t, "mailjet-key", cfg.MailjetAPIKey)
	assert.Equal(t, "mailjet-secret", cfg.MailjetAPISecret)
	assert.Equal(t, "webhook-secret", cfg.MailjetWebhookSecret)
	assert.Equal(t, "sender@example.com", cfg.SenderEmail)
	assert.Equal(t, "sends-table", cfg.SendsTable)
	assert.Equal(t, "profiles-table", cfg.UserProfileTable)
	assert.Equal(t, consts.LoggingBackendSentry, cfg.LoggingProvider)
	assert.Equal(t, "sentry-dsn", cfg.SentryDSN)
	assert.Equal(t, "production", cfg.SentryEnvironment)
	assert.Equal(t, 0.5, cfg.SentrySampleRate)
}

func TestLoad_Boolean_String_Conversion(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"true as string", "true", true},
		{"TRUE as string", "TRUE", true},
		{"1 as string", "1", true},
		{"false as string", "false", false},
		{"FALSE as string", "FALSE", false},
		{"0 as string", "0", false},
		{"empty string defaults to false", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupEnvVars(t, map[string]string{
				"SAVETOINK_DEBUG": tt.value,
			})

			cfg, err := Load(consts.ModeCLI, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg.Debug)
		})
	}
}

func TestLoad_Float_Conversion(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected float64
	}{
		{"integer", "1", 1.0},
		{"float", "0.5", 0.5},
		{"zero", "0", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupEnvVars(t, map[string]string{
				"SAVETOINK_API_KEY":                 "test-api-key",
				"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
				"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
				"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
				"SAVETOINK_APP_URL":                 "https://example.com",
				"SAVETOINK_SENTRY_SAMPLE_RATE":      tt.value,
			})

			awsLoader := mockAWSLoader()

			cfg, err := Load(consts.ModeServer, awsLoader)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg.SentrySampleRate)
		})
	}
}

func mockAWSLoader() AWSConfigLoader {
	return func(_ context.Context) (aws.Config, error) {
		return aws.Config{}, nil
	}
}

func setupEnvVars(t *testing.T, envVars map[string]string) {
	t.Helper()

	oldEnv := make(map[string]string)

	for key, value := range envVars {
		oldValue, exists := os.LookupEnv(key)
		if exists {
			oldEnv[key] = oldValue
		}
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("failed to set env var %s: %v", key, err)
		}
	}

	t.Cleanup(func() {
		for key := range envVars {
			if oldValue, exists := oldEnv[key]; exists {
				if err := os.Setenv(key, oldValue); err != nil {
					t.Errorf("failed to restore env var %s: %v", key, err)
				}
			} else {
				if err := os.Unsetenv(key); err != nil {
					t.Errorf("failed to unset env var %s: %v", key, err)
				}
			}
		}
	})
}
