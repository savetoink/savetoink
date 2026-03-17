package config

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	_ = os.Unsetenv("SAVETOINK_API_KEY")
	_ = os.Unsetenv("SAVETOINK_ARTICLE_TABLE_NAME")
	_ = os.Unsetenv("SAVETOINK_USER_PROFILE_TABLE_NAME")
	_ = os.Unsetenv("SAVETOINK_SENDS_TABLE_NAME")
	_ = os.Unsetenv("SAVETOINK_APP_URL")
	_ = os.Unsetenv("SAVETOINK_CORS_ALLOW_ORIGIN")
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
	_ = os.Unsetenv("SAVETOINK_STORAGE_BACKEND")
	_ = os.Unsetenv("SAVETOINK_SQLITE_PATH")
	_ = os.Unsetenv("SAVETOINK_BROWSERLESS_KEY")
	_ = os.Unsetenv("SAVETOINK_PROCESS_ARTICLE_LAMBDA")
	_ = os.Unsetenv("SAVETOINK_TASKS")
	_ = os.Unsetenv("SAVETOINK_HTTP_PORT")
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
		"SAVETOINK_STORAGE_BACKEND":         "dynamodb",
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
		"SAVETOINK_STORAGE_BACKEND":         "dynamodb",
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
		"SAVETOINK_STORAGE_BACKEND":         "dynamodb",
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
		"SAVETOINK_AUTH_BACKEND":    "invalid-backend",
		"SAVETOINK_STORAGE_BACKEND": "sqlite",
		"SAVETOINK_SQLITE_PATH":     "/path/to/database.db",
		"SAVETOINK_APP_URL":         "https://example.com",
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
		"SAVETOINK_STORAGE_BACKEND": "sqlite",
		"SAVETOINK_SQLITE_PATH":     "/path/to/database.db",
		"SAVETOINK_APP_URL":         "https://example.com",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SAVETOINK_API_KEY")
}

func TestLoad_Missing_Multiple_Auth0_Env(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_AUTH_BACKEND":    "auth0",
		"SAVETOINK_STORAGE_BACKEND": "sqlite",
		"SAVETOINK_SQLITE_PATH":     "/path/to/database.db",
		"SAVETOINK_APP_URL":         "https://example.com",
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
		"SAVETOINK_STORAGE_BACKEND":         "dynamodb",
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SAVETOINK_ARTICLE_TABLE_NAME")
}

func TestLoad_Missing_UserProfileTable(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND":    "dynamodb",
		"SAVETOINK_API_KEY":            "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME": "articles-table",
		"SAVETOINK_SENDS_TABLE_NAME":   "sends-table",
		"SAVETOINK_APP_URL":            "https://example.com",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SAVETOINK_USER_PROFILE_TABLE_NAME")
}

func TestLoad_Missing_SendsTable(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND":    "dynamodb",
		"SAVETOINK_API_KEY":            "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME": "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE": "profiles-table",
		"SAVETOINK_APP_URL":            "https://example.com",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SAVETOINK_SENDS_TABLE_NAME")
}

func TestLoad_Missing_AppURL(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND":    "sqlite",
		"SAVETOINK_SQLITE_PATH":        "/path/to/database.db",
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
		"SAVETOINK_STORAGE_BACKEND":    "sqlite",
		"SAVETOINK_SQLITE_PATH":        "/path/to/database.db",
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
		"SAVETOINK_STORAGE_BACKEND":    "sqlite",
		"SAVETOINK_SQLITE_PATH":        "/path/to/database.db",
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
		"SAVETOINK_STORAGE_BACKEND":    "sqlite",
		"SAVETOINK_SQLITE_PATH":        "/path/to/database.db",
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
		"SAVETOINK_STORAGE_BACKEND":     "sqlite",
		"SAVETOINK_SQLITE_PATH":         "/path/to/database.db",
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
		"SAVETOINK_STORAGE_BACKEND":         "sqlite",
		"SAVETOINK_SQLITE_PATH":             "/path/to/database.db",
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
		"SAVETOINK_STORAGE_BACKEND":    "sqlite",
		"SAVETOINK_SQLITE_PATH":        "/path/to/database.db",
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
		"SAVETOINK_STORAGE_BACKEND":    "sqlite",
		"SAVETOINK_SQLITE_PATH":        "/path/to/database.db",
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
		"SAVETOINK_STORAGE_BACKEND":    "sqlite",
		"SAVETOINK_SQLITE_PATH":        "/path/to/database.db",
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
		"SAVETOINK_STORAGE_BACKEND":         "sqlite",
		"SAVETOINK_SQLITE_PATH":             "/path/to/database.db",
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
		"SAVETOINK_AUTH_BACKEND":    "auth0",
		"SAVETOINK_STORAGE_BACKEND": "dynamodb",
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
		"SAVETOINK_STORAGE_BACKEND":         "dynamodb",
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
	})

	_, err := Load(consts.ModeServer, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AWS config loader is required")
}

func TestLoad_Server_Mode_AWS_Loader_Error(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND":         "dynamodb",
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
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
		"SAVETOINK_CORS_ALLOW_ORIGIN":       "https://example.com",
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
		"SAVETOINK_STORAGE_BACKEND":         "dynamodb",
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
	assert.Equal(t, "https://example.com", cfg.CorsAllowOrigin)
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

func TestLoad_CorsAllowOrigin(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"specific origin", "https://example.com", "https://example.com"},
		{"localhost", "http://localhost:3000", "http://localhost:3000"},
		{"wildcard", "*", "*"},
		{"empty string", "", "*"},
		{"multiple origins (as-is)", "https://example.com,https://test.com", "https://example.com,https://test.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupEnvVars(t, map[string]string{
				"SAVETOINK_STORAGE_BACKEND":         "sqlite",
				"SAVETOINK_SQLITE_PATH":             "/path/to/database.db",
				"SAVETOINK_API_KEY":                 "test-api-key",
				"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
				"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
				"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
				"SAVETOINK_APP_URL":                 "https://example.com",
				"SAVETOINK_CORS_ALLOW_ORIGIN":       tt.value,
			})

			awsLoader := mockAWSLoader()

			cfg, err := Load(consts.ModeServer, awsLoader)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg.CorsAllowOrigin)
		})
	}
}

func TestLoad_SQLite_Backend_Success(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND": "sqlite",
		"SAVETOINK_SQLITE_PATH":     "/path/to/database.db",
		"SAVETOINK_APP_URL":         "https://example.com",
		"SAVETOINK_API_KEY":         "test-api-key",
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.StorageBackendSQLite, cfg.StorageBackend)
	assert.Equal(t, "/path/to/database.db", cfg.SQLitePath)
}

func TestLoad_DynamoDB_Backend_Success(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND":         "dynamodb",
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
	})

	awsLoader := mockAWSLoader()

	cfg, err := Load(consts.ModeServer, awsLoader)
	require.NoError(t, err)
	assert.Equal(t, consts.StorageBackendDynamoDB, cfg.StorageBackend)
	assert.NotNil(t, cfg.AWSConfig)
}

func TestLoad_Default_Storage_Backend(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_API_KEY":     "test-api-key",
		"SAVETOINK_SQLITE_PATH": "/path/to/database.db",
		"SAVETOINK_APP_URL":     "https://example.com",
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.StorageBackendSQLite, cfg.StorageBackend)
}

func TestLoad_SQLite_Path_Default(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND": "sqlite",
		"SAVETOINK_APP_URL":         "https://example.com",
		"SAVETOINK_API_KEY":         "test-api-key",
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.StorageBackendSQLite, cfg.StorageBackend)
	assert.Equal(t, consts.SQLitePathDefault, cfg.SQLitePath)
}

func TestLoad_Missing_DynamoDB_Tables(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND": "dynamodb",
		"SAVETOINK_API_KEY":         "test-api-key",
		"SAVETOINK_APP_URL":         "https://example.com",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	errMsg := err.Error()
	assert.Contains(t, errMsg, "SAVETOINK_ARTICLE_TABLE_NAME")
	assert.Contains(t, errMsg, "SAVETOINK_USER_PROFILE_TABLE_NAME")
	assert.Contains(t, errMsg, "SAVETOINK_SENDS_TABLE_NAME")
}

func TestLoad_Invalid_Storage_Backend(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND": "invalid-backend",
		"SAVETOINK_APP_URL":         "https://example.com",
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported storage backend")
}

func TestLoad_Browserless_Key(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND":         "sqlite",
		"SAVETOINK_SQLITE_PATH":             "/path/to/database.db",
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
		"SAVETOINK_BROWSERLESS_KEY":         "browserless-key-123",
	})

	awsLoader := mockAWSLoader()

	cfg, err := Load(consts.ModeServer, awsLoader)
	require.NoError(t, err)
	assert.Equal(t, "browserless-key-123", cfg.BrowserlessKey)
}

func TestLoad_Process_Article_Lambda(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND":         "sqlite",
		"SAVETOINK_SQLITE_PATH":             "/path/to/database.db",
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
		"SAVETOINK_PROCESS_ARTICLE_LAMBDA":  "arn:aws:lambda:us-east-1:123456789012:function:ProcessArticle",
	})

	awsLoader := mockAWSLoader()

	cfg, err := Load(consts.ModeServer, awsLoader)
	require.NoError(t, err)
	assert.Equal(t, "arn:aws:lambda:us-east-1:123456789012:function:ProcessArticle", cfg.ProcessArticleLambda)
}

func TestLoad_CLI_WithMailjet(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_EMAIL_BACKEND":      "mailjet",
		"SAVETOINK_MAILJET_API_KEY":    "mailjet-key",
		"SAVETOINK_MAILJET_API_SECRET": "mailjet-secret",
		"SAVETOINK_SENDER_EMAIL":       "sender@example.com",
	})

	cfg, err := Load(consts.ModeCLI, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.EmailBackendMailjet, cfg.EmailProvider)
	assert.Equal(t, "mailjet-key", cfg.MailjetAPIKey)
	assert.Equal(t, "mailjet-secret", cfg.MailjetAPISecret)
	assert.Equal(t, "sender@example.com", cfg.SenderEmail)
}

func TestLoad_InvalidEmailProvider(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_EMAIL_BACKEND": "invalid-provider",
	})

	cfg, err := Load(consts.ModeCLI, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.EmailProvider("invalid-provider"), cfg.EmailProvider)
}

func TestLoad_MissingMailjetKeys_CLI(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_EMAIL_BACKEND": "mailjet",
	})

	cfg, err := Load(consts.ModeCLI, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.EmailBackendMailjet, cfg.EmailProvider)
	assert.Empty(t, cfg.MailjetAPIKey)
	assert.Empty(t, cfg.MailjetAPISecret)
	assert.Empty(t, cfg.SenderEmail)
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

func TestLoad_Tasks_Config(t *testing.T) {
	tests := []struct {
		name     string
		tasksEnv string
		expected []string
	}{
		{
			name:     "single task",
			tasksEnv: `[{"task":"task1","schedule":"0 * * * *"}]`,
			expected: []string{"task1"},
		},
		{
			name: "multiple tasks",
			tasksEnv: `[{"task":"task1","schedule":"0 * * * *"},` +
				`{"task":"task2","schedule":"0 2 * * *"}]`,
			expected: []string{"task1", "task2"},
		},
		{
			name:     "empty tasks",
			tasksEnv: `[]`,
			expected: []string{},
		},
		{
			name:     "no tasks env",
			tasksEnv: "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars := map[string]string{
				"SAVETOINK_STORAGE_BACKEND":         "sqlite",
				"SAVETOINK_SQLITE_PATH":             "/path/to/database.db",
				"SAVETOINK_API_KEY":                 "test-api-key",
				"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
				"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
				"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
				"SAVETOINK_APP_URL":                 "https://example.com",
			}
			if tt.tasksEnv != "" {
				envVars["SAVETOINK_TASKS"] = tt.tasksEnv
			}

			setupEnvVars(t, envVars)

			cfg, err := Load(consts.ModeServer, nil)
			require.NoError(t, err)

			if tt.expected == nil {
				assert.Nil(t, cfg.Tasks)
			} else {
				require.Len(t, cfg.Tasks, len(tt.expected))
				for i, name := range tt.expected {
					assert.Equal(t, name, cfg.Tasks[i].Task)
				}
			}
		})
	}
}

func TestLoad_DefaultPort(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND":         "sqlite",
		"SAVETOINK_SQLITE_PATH":             "/path/to/database.db",
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.DefaultHTTPPort, cfg.Port)
}

func TestLoad_CustomPort(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND":         "sqlite",
		"SAVETOINK_SQLITE_PATH":             "/path/to/database.db",
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
		"SAVETOINK_HTTP_PORT":               "3000",
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	assert.Equal(t, 3000, cfg.Port)
}

func TestLoad_CLI_Mode_DefaultPort(t *testing.T) {
	setupEnvVars(t, map[string]string{})

	cfg, err := Load(consts.ModeCLI, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.DefaultHTTPPort, cfg.Port)
}

func TestLoad_CLI_Mode_CustomPort(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_HTTP_PORT": "9000",
	})

	cfg, err := Load(consts.ModeCLI, nil)
	require.NoError(t, err)
	assert.Equal(t, 9000, cfg.Port)
}

func TestLoad_CLI_InvalidStorageBackend(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND": "invalid-backend",
		"SAVETOINK_APP_URL":         "https://example.com",
		"SAVETOINK_API_KEY":         "test-api-key",
	})

	cfg, err := Load(consts.ModeCLI, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.StorageBackend("invalid-backend"), cfg.StorageBackend)
}

func TestLoad_CLI_EmptyStorageBackend(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND": "",
		"SAVETOINK_APP_URL":         "https://example.com",
		"SAVETOINK_API_KEY":         "test-api-key",
	})

	cfg, err := Load(consts.ModeCLI, nil)
	require.NoError(t, err)
	assert.Empty(t, cfg.StorageBackend)
}

func TestLoad_Server_SQLite_WithoutPath(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND":    "sqlite",
		"SAVETOINK_APP_URL":            "https://example.com",
		"SAVETOINK_API_KEY":            "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME": "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":   "sends-table",
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.StorageBackendSQLite, cfg.StorageBackend)
	assert.Equal(t, consts.SQLitePathDefault, cfg.SQLitePath)
}

func TestLoad_Tasks_Config_InvalidJSON(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND":         "sqlite",
		"SAVETOINK_SQLITE_PATH":             "/path/to/database.db",
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
		"SAVETOINK_TASKS":                   "invalid-json{",
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	assert.Nil(t, cfg.Tasks)
}

func TestLoad_ArticleTagsTable(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND":         "sqlite",
		"SAVETOINK_SQLITE_PATH":             "/path/to/database.db",
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_ARTICLE_TAGS_TABLE_NAME": "article-tags-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	assert.Equal(t, "article-tags-table", cfg.ArticleTagsTable)
}

func TestLoad_BackupBucketName(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND":         "sqlite",
		"SAVETOINK_SQLITE_PATH":             "/path/to/database.db",
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
		"SAVETOINK_BACKUP_BUCKET_NAME":      "backup-bucket",
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	assert.Equal(t, "backup-bucket", cfg.BackupBucketName)
}

func TestValidateSQLiteConfig_EmptyPath(t *testing.T) {
	cfg := &Config{SQLitePath: ""}
	var missing []string

	err := cfg.validateSQLiteConfig(&missing)
	require.NoError(t, err)
	assert.Contains(t, missing, "SAVETOINK_SQLITE_PATH")
}

func TestValidateStorageBackendConfig_EmptyBackend(t *testing.T) {
	cfg := &Config{StorageBackend: "", SQLitePath: "/path/to/db.db"}
	var missing []string

	err := cfg.validateStorageBackendConfig(&missing, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.StorageBackendSQLite, cfg.StorageBackend)
}

func TestValidateEmailProviderConfigCli_Mailjet(t *testing.T) {
	cfg := &Config{
		EmailProvider: consts.EmailBackendMailjet,
	}
	var missing []string

	cfg.ValidateEmailProviderConfigCli(&missing)
	assert.Contains(t, missing, "SAVETOINK_MAILJET_API_KEY")
	assert.Contains(t, missing, "SAVETOINK_MAILJET_API_SECRET")
	assert.Contains(t, missing, "SAVETOINK_SENDER_EMAIL")
}

func TestValidateEmailProviderConfigCli_NonMailjet(t *testing.T) {
	cfg := &Config{
		EmailProvider: consts.EmailProvider("other-provider"),
	}
	var missing []string

	cfg.ValidateEmailProviderConfigCli(&missing)
	assert.Empty(t, missing)
}

func TestValidateServerConfig_SQLite(t *testing.T) {
	cfg := &Config{
		StorageBackend: consts.StorageBackendSQLite,
		SQLitePath:     "/path/to/db.db",
		AuthBackend:    consts.AuthBackendSharedAPIKey,
		APIKeySecret:   "test-key",
		AppURL:         "https://example.com",
	}
	var missing []string

	err := cfg.validateServerConfig(&missing, nil)
	require.NoError(t, err)
}

func TestValidateServerConfig_DynamoDB(t *testing.T) {
	cfg := &Config{
		StorageBackend:   consts.StorageBackendDynamoDB,
		ArticlesTable:    "articles",
		UserProfileTable: "profiles",
		SendsTable:       "sends",
		AuthBackend:      consts.AuthBackendSharedAPIKey,
		APIKeySecret:     "test-key",
		AppURL:           "https://example.com",
	}
	var missing []string

	awsLoader := mockAWSLoader()

	err := cfg.validateServerConfig(&missing, awsLoader)
	require.NoError(t, err)
}

func TestLoad_Tasks_Config_InvalidTask(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND":         "sqlite",
		"SAVETOINK_SQLITE_PATH":             "/path/to/database.db",
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
		"SAVETOINK_TASKS":                   `[{"task":"task1"}]`,
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	require.Len(t, cfg.Tasks, 1)
	assert.Equal(t, "task1", cfg.Tasks[0].Task)
	assert.Empty(t, cfg.Tasks[0].Schedule)
}

func TestLoad_Tasks_Config_WithBackup(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND":         "sqlite",
		"SAVETOINK_SQLITE_PATH":             "/path/to/database.db",
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
		"SAVETOINK_TASKS":                   `[{"task":"backup","schedule":"0 * * * *"}]`,
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	require.Len(t, cfg.Tasks, 1)
	assert.Equal(t, "backup", cfg.Tasks[0].Task)
	assert.Equal(t, "0 * * * *", cfg.Tasks[0].Schedule)
}

func TestLoad_Tasks_Config_WithRestore(t *testing.T) {
	setupEnvVars(t, map[string]string{
		"SAVETOINK_STORAGE_BACKEND":         "sqlite",
		"SAVETOINK_SQLITE_PATH":             "/path/to/database.db",
		"SAVETOINK_API_KEY":                 "test-api-key",
		"SAVETOINK_ARTICLE_TABLE_NAME":      "articles-table",
		"SAVETOINK_USER_PROFILE_TABLE_NAME": "profiles-table",
		"SAVETOINK_SENDS_TABLE_NAME":        "sends-table",
		"SAVETOINK_APP_URL":                 "https://example.com",
		"SAVETOINK_TASKS":                   `[{"task":"restore","backup_name":"backup-123","overwrite":true}]`,
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	require.Len(t, cfg.Tasks, 1)
	assert.Equal(t, "restore", cfg.Tasks[0].Task)
	assert.Equal(t, "backup-123", cfg.Tasks[0].BackupName)
	assert.True(t, cfg.Tasks[0].Overwrite)
}
