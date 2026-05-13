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

const (
	testSQLitePath    = "/path/to/database.db"
	testSQLitePathAlt = "/path/to/db.db"
	testPasetoKey     = "QUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUE="
	testSentryRate    = "1.0"

	testArticlesTable     = "articles-table"
	testProfilesTable     = "profiles-table"
	testSendsTable        = "sends-table"
	testMailjetAPIKey     = "mailjet-key"
	testMailjetAPISecret  = "mailjet-secret"
	testAPIKeySecret      = "test-api-key"
	testSenderEmail       = "sender@example.com"
	testAuth0Domain       = "auth0-domain"
	testAuth0Audience     = "auth0-audience"
	testAuth0ClientID     = "auth0-client-id"
	testAuth0ClientSecret = "auth0-client-secret" //nolint:gosec // test value, not a real credential
	testSentryDSN         = "sentry-dsn"
	testSentryEnv         = "production"
	testInvalidBackend    = "invalid-backend"
	testCorsAllowOrigin   = "https://example.com"
	testEnvProfileTable   = "SAVETOINK_USER_PROFILE_TABLE"
	testTrue              = "true"
	testWebhookSecret     = "webhook-secret"
	testSQLite            = "sqlite"
	testFalse             = "false"
	testV1                = "v1"
	testDynamoDB          = "dynamodb"
	testLocalhost3000     = "http://localhost:3000"
	testMultiOrigins      = "https://example.com,https://test.com"
	testTask1             = "task1"
	testAPIKey            = "test-key"
	testAuth0Str          = "auth0"
)

func TestMain(m *testing.M) {
	_ = os.Unsetenv(envAPIKeySecret)
	_ = os.Unsetenv(envArticleTableName)
	_ = os.Unsetenv(envUserProfileTableName)
	_ = os.Unsetenv(envSendsTableName)
	_ = os.Unsetenv(envCorsAllowOrigin)
	_ = os.Unsetenv(envEmailBackend)
	_ = os.Unsetenv(envMailjetAPIKey)
	_ = os.Unsetenv(envMailjetAPISecret)
	_ = os.Unsetenv(envMailjetWebhookSecret)
	_ = os.Unsetenv(envSenderEmail)
	_ = os.Unsetenv(envAuthBackend)
	_ = os.Unsetenv(envAuth0Domain)
	_ = os.Unsetenv(envAuth0Audience)
	_ = os.Unsetenv(envAuth0ClientID)
	_ = os.Unsetenv(envAuth0ClientSecret)
	_ = os.Unsetenv(envLoggingProvider)
	_ = os.Unsetenv(envSentryDSN)
	_ = os.Unsetenv(envSentryEnvironment)
	_ = os.Unsetenv(envSentrySampleRate)
	_ = os.Unsetenv(envDebug)
	_ = os.Unsetenv(envDisableQuotaCheck)
	_ = os.Unsetenv(envStorageBackend)
	_ = os.Unsetenv(envSQLitePath)
	_ = os.Unsetenv(envBrowserlessKey)
	_ = os.Unsetenv(envProcessArticleLambda)
	_ = os.Unsetenv(envTasks)
	_ = os.Unsetenv(envHTTPPort)
	_ = os.Unsetenv(envPasetoKey)
	_ = os.Unsetenv(envPasetoKeyVersion)
	os.Exit(m.Run())
}

func TestLoad_CLI_Mode(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envDebug: testFalse,
	})

	cfg, err := Load(consts.ModeCLI, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.ModeCLI, cfg.Mode)
	assert.False(t, cfg.Debug)
	assert.Nil(t, cfg.AWSConfig)
}

func TestLoad_CLI_Mode_With_Debug(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envDebug: testTrue,
	})

	cfg, err := Load(consts.ModeCLI, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.ModeCLI, cfg.Mode)
	assert.True(t, cfg.Debug)
}

func TestLoad_CLI_Mode_DisableQuotaCheck(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"quota check disabled", testTrue, true},
		{"quota check enabled", testFalse, false},
		{"quota check enabled with empty string", "", false},
		{"quota check enabled with invalid value", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupEnvVars(t, map[string]string{
				envDisableQuotaCheck: tt.value,
			})

			cfg, err := Load(consts.ModeCLI, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg.DisableQuotaCheck)
		})
	}
}

func TestLoad_Server_Mode_DisableQuotaCheck(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendSQLite),
		envSQLitePath:           testSQLitePath,
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
		envDisableQuotaCheck:    testTrue,
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	assert.True(t, cfg.DisableQuotaCheck)
}

func TestLoad_Server_Mode_Success(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendDynamoDB),
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
		envEmailBackend:         string(consts.EmailBackendMailjet),
		envMailjetAPIKey:        testMailjetAPIKey,
		envMailjetAPISecret:     testMailjetAPISecret,
		envMailjetWebhookSecret: testWebhookSecret,
		envSenderEmail:          testSenderEmail,
	})

	awsLoader := mockAWSLoader()

	cfg, err := Load(consts.ModeServer, awsLoader)
	require.NoError(t, err)
	assert.Equal(t, consts.ModeServer, cfg.Mode)
	assert.NotNil(t, cfg.AWSConfig)
	assert.Equal(t, testAPIKeySecret, cfg.APIKeySecret)
	assert.Equal(t, testArticlesTable, cfg.ArticlesTable)
	assert.Equal(t, testProfilesTable, cfg.UserProfileTable)
	assert.Equal(t, testSendsTable, cfg.SendsTable)
	assert.Equal(t, consts.EmailBackendMailjet, cfg.EmailProvider)
}

func TestLoad_Default_Auth_Backend(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendDynamoDB),
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
	})

	awsLoader := mockAWSLoader()

	cfg, err := Load(consts.ModeServer, awsLoader)
	require.NoError(t, err)
	assert.Equal(t, consts.AuthBackendSharedAPIKey, cfg.AuthBackend)
}

func TestLoad_Auth0_Backend_Success(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envAuthBackend:          string(consts.AuthBackendAuth0),
		envAuth0Domain:          testAuth0Domain,
		envAuth0Audience:        testAuth0Audience,
		envAuth0ClientID:        testAuth0ClientID,
		envAuth0ClientSecret:    testAuth0ClientSecret,
		envPasetoKey:            testPasetoKey,
		envPasetoKeyVersion:     testV1,
		envStorageBackend:       string(consts.StorageBackendDynamoDB),
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
	})

	awsLoader := mockAWSLoader()

	cfg, err := Load(consts.ModeServer, awsLoader)
	require.NoError(t, err)
	assert.Equal(t, consts.AuthBackendAuth0, cfg.AuthBackend)
	assert.Equal(t, testAuth0Domain, cfg.Auth0Domain)
	assert.Equal(t, testAuth0Audience, cfg.Auth0Audience)
	assert.Equal(t, testAuth0ClientID, cfg.Auth0ClientID)
	assert.Equal(t, testAuth0ClientSecret, cfg.Auth0ClientSecret)
	assert.Equal(t, testPasetoKey, cfg.PASETOSymmetricKey)
	assert.Equal(t, testV1, cfg.PASETOKeyVersion)
}

func TestLoad_Invalid_Auth_Backend(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envAuthBackend:    testInvalidBackend,
		envStorageBackend: string(consts.StorageBackendSQLite),
		envSQLitePath:     testSQLitePath,
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
		envStorageBackend: string(consts.StorageBackendSQLite),
		envSQLitePath:     testSQLitePath,
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), envAPIKeySecret)
}

func TestLoad_Missing_Multiple_Auth0_Env(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envAuthBackend:    string(consts.AuthBackendAuth0),
		envStorageBackend: string(consts.StorageBackendSQLite),
		envSQLitePath:     testSQLitePath,
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	errMsg := err.Error()
	assert.Contains(t, errMsg, envAuth0Domain)
	assert.Contains(t, errMsg, envAuth0Audience)
	assert.Contains(t, errMsg, envAuth0ClientID)
	assert.Contains(t, errMsg, envAuth0ClientSecret)
	assert.Contains(t, errMsg, envPasetoKey)
	assert.Contains(t, errMsg, envPasetoKeyVersion)
}

func TestLoad_Missing_ArticlesTable(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendDynamoDB),
		envAPIKeySecret:         testAPIKeySecret,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), envArticleTableName)
}

func TestLoad_Missing_UserProfileTable(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:   testDynamoDB,
		envAPIKeySecret:     testAPIKeySecret,
		envArticleTableName: testArticlesTable,
		envSendsTableName:   testSendsTable,
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), envUserProfileTableName)
}

func TestLoad_Missing_SendsTable(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:   testDynamoDB,
		envAPIKeySecret:     testAPIKeySecret,
		envArticleTableName: testArticlesTable,
		testEnvProfileTable: testProfilesTable,
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), envSendsTableName)
}

func TestLoad_Missing_Mailjet_API_Key(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:   string(consts.StorageBackendSQLite),
		envSQLitePath:       testSQLitePath,
		envAPIKeySecret:     testAPIKeySecret,
		envArticleTableName: testArticlesTable,
		testEnvProfileTable: testProfilesTable,
		envSendsTableName:   testSendsTable,
		envEmailBackend:     string(consts.EmailBackendMailjet),
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), envMailjetAPIKey)
}

func TestLoad_Missing_Mailjet_API_Secret(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:   string(consts.StorageBackendSQLite),
		envSQLitePath:       testSQLitePath,
		envAPIKeySecret:     testAPIKeySecret,
		envArticleTableName: testArticlesTable,
		testEnvProfileTable: testProfilesTable,
		envSendsTableName:   testSendsTable,
		envEmailBackend:     string(consts.EmailBackendMailjet),
		envMailjetAPIKey:    testMailjetAPIKey,
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), envMailjetAPISecret)
}

func TestLoad_Missing_Mailjet_Webhook_Secret(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:   string(consts.StorageBackendSQLite),
		envSQLitePath:       testSQLitePath,
		envAPIKeySecret:     testAPIKeySecret,
		envArticleTableName: testArticlesTable,
		testEnvProfileTable: testProfilesTable,
		envSendsTableName:   testSendsTable,
		envEmailBackend:     string(consts.EmailBackendMailjet),
		envMailjetAPIKey:    testMailjetAPIKey,
		envMailjetAPISecret: testMailjetAPISecret,
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), envMailjetWebhookSecret)
}

func TestLoad_Missing_Sender_Email(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:               testSQLite,
		envSQLitePath:                   testSQLitePath,
		envAPIKeySecret:                 testAPIKeySecret,
		envArticleTableName:             testArticlesTable,
		testEnvProfileTable:             testProfilesTable,
		envSendsTableName:               testSendsTable,
		envEmailBackend:                 "mailjet",
		envMailjetAPIKey:                testMailjetAPIKey,
		envMailjetAPISecret:             testMailjetAPISecret,
		"SAVETOINK_MAILJET_WEBHOOK_SEC": testWebhookSecret,
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), envSenderEmail)
}

func TestLoad_Sentry_Provider_Success(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendSQLite),
		envSQLitePath:           testSQLitePath,
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
		envLoggingProvider:      string(consts.LoggingBackendSentry),
		envSentryDSN:            testSentryDSN,
		envSentryEnvironment:    testSentryEnv,
		envSentrySampleRate:     testSentryRate,
	})

	awsLoader := mockAWSLoader()

	cfg, err := Load(consts.ModeServer, awsLoader)
	require.NoError(t, err)
	assert.Equal(t, consts.LoggingBackendSentry, cfg.LoggingProvider)
	assert.Equal(t, testSentryDSN, cfg.SentryDSN)
	assert.Equal(t, testSentryEnv, cfg.SentryEnvironment)
	assert.Equal(t, 1.0, cfg.SentrySampleRate)
}

func TestLoad_Missing_Sentry_DSN(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:    string(consts.StorageBackendSQLite),
		envSQLitePath:        testSQLitePath,
		envAPIKeySecret:      testAPIKeySecret,
		envArticleTableName:  testArticlesTable,
		testEnvProfileTable:  testProfilesTable,
		envSendsTableName:    testSendsTable,
		envLoggingProvider:   string(consts.LoggingBackendSentry),
		envSentryEnvironment: testSentryEnv,
		envSentrySampleRate:  testSentryRate,
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), envSentryDSN)
}

func TestLoad_Missing_Sentry_Environment(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:   string(consts.StorageBackendSQLite),
		envSQLitePath:       testSQLitePath,
		envAPIKeySecret:     testAPIKeySecret,
		envArticleTableName: testArticlesTable,
		testEnvProfileTable: testProfilesTable,
		envSendsTableName:   testSendsTable,
		envLoggingProvider:  string(consts.LoggingBackendSentry),
		envSentryDSN:        testSentryDSN,
		envSentrySampleRate: testSentryRate,
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), envSentryEnvironment)
}

func TestLoad_Missing_Sentry_Sample_Rate(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:    string(consts.StorageBackendSQLite),
		envSQLitePath:        testSQLitePath,
		envAPIKeySecret:      testAPIKeySecret,
		envArticleTableName:  testArticlesTable,
		testEnvProfileTable:  testProfilesTable,
		envSendsTableName:    testSendsTable,
		envLoggingProvider:   string(consts.LoggingBackendSentry),
		envSentryDSN:         testSentryDSN,
		envSentryEnvironment: testSentryEnv,
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), envSentrySampleRate)
}

func TestLoad_Non_Sentry_Logging_Does_Not_Validate_Sentry(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendSQLite),
		envSQLitePath:           testSQLitePath,
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
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
		envAuthBackend:    string(consts.AuthBackendAuth0),
		envStorageBackend: string(consts.StorageBackendDynamoDB),
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	errMsg := err.Error()
	assert.Contains(t, errMsg, envAuth0Domain)
	assert.Contains(t, errMsg, envAuth0Audience)
	assert.Contains(t, errMsg, envAuth0ClientID)
	assert.Contains(t, errMsg, envAuth0ClientSecret)
	assert.Contains(t, errMsg, envPasetoKey)
	assert.Contains(t, errMsg, envPasetoKeyVersion)
	assert.Contains(t, errMsg, envArticleTableName)
	assert.Contains(t, errMsg, envUserProfileTableName)
	assert.Contains(t, errMsg, envSendsTableName)
}

func TestLoad_Server_Mode_Without_AWS_Loader(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendDynamoDB),
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
	})

	_, err := Load(consts.ModeServer, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AWS config loader is required")
}

func TestLoad_Server_Mode_AWS_Loader_Error(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendDynamoDB),
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
	})

	awsLoader := func(_ context.Context) (aws.Config, error) {
		return aws.Config{}, assert.AnError
	}

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load AWS config")
}

func TestLoad_All_Env_Vars_Bound(t *testing.T) {
	envVars := map[string]string{
		envAPIKeySecret:         "api-key-secret",
		envArticleTableName:     testArticlesTable,
		envAuth0Audience:        testAuth0Audience,
		envAuth0ClientID:        testAuth0ClientID,
		envAuth0ClientSecret:    testAuth0ClientSecret,
		envAuth0Domain:          testAuth0Domain,
		envAuthBackend:          "shared_api_key",
		envCorsAllowOrigin:      testCorsAllowOrigin,
		envDebug:                testTrue,
		envEmailBackend:         string(consts.EmailBackendMailjet),
		envMailjetAPIKey:        testMailjetAPIKey,
		envMailjetAPISecret:     testMailjetAPISecret,
		envMailjetWebhookSecret: testWebhookSecret,
		envSenderEmail:          testSenderEmail,
		envSendsTableName:       testSendsTable,
		envUserProfileTableName: testProfilesTable,
		envLoggingProvider:      string(consts.LoggingBackendSentry),
		envSentryDSN:            testSentryDSN,
		envSentryEnvironment:    testSentryEnv,
		envSentrySampleRate:     "0.5",
		envStorageBackend:       string(consts.StorageBackendDynamoDB),
	}

	setupEnvVars(t, envVars)

	awsLoader := mockAWSLoader()

	cfg, err := Load(consts.ModeServer, awsLoader)
	require.NoError(t, err)

	assert.Equal(t, "api-key-secret", cfg.APIKeySecret)
	assert.Equal(t, testArticlesTable, cfg.ArticlesTable)
	assert.Equal(t, testAuth0Audience, cfg.Auth0Audience)
	assert.Equal(t, testAuth0ClientID, cfg.Auth0ClientID)
	assert.Equal(t, testAuth0ClientSecret, cfg.Auth0ClientSecret)
	assert.Equal(t, testAuth0Domain, cfg.Auth0Domain)
	assert.Equal(t, consts.AuthBackendSharedAPIKey, cfg.AuthBackend)
	assert.Equal(t, testCorsAllowOrigin, cfg.CorsAllowOrigin)
	assert.True(t, cfg.Debug)
	assert.Equal(t, consts.EmailBackendMailjet, cfg.EmailProvider)
	assert.Equal(t, testMailjetAPIKey, cfg.MailjetAPIKey)
	assert.Equal(t, testMailjetAPISecret, cfg.MailjetAPISecret)
	assert.Equal(t, testWebhookSecret, cfg.MailjetWebhookSecret)
	assert.Equal(t, testSenderEmail, cfg.SenderEmail)
	assert.Equal(t, testSendsTable, cfg.SendsTable)
	assert.Equal(t, testProfilesTable, cfg.UserProfileTable)
	assert.Equal(t, consts.LoggingBackendSentry, cfg.LoggingProvider)
	assert.Equal(t, testSentryDSN, cfg.SentryDSN)
	assert.Equal(t, testSentryEnv, cfg.SentryEnvironment)
	assert.Equal(t, 0.5, cfg.SentrySampleRate)
}

func TestLoad_CorsAllowOrigin(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"specific origin", testCorsAllowOrigin, testCorsAllowOrigin},
		{"localhost", testLocalhost3000, testLocalhost3000},
		{"wildcard", "*", "*"},
		{"empty string", "", "*"},
		{"multiple origins (as-is)", testMultiOrigins, testMultiOrigins},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupEnvVars(t, map[string]string{
				envStorageBackend:       string(consts.StorageBackendSQLite),
				envSQLitePath:           testSQLitePath,
				envAPIKeySecret:         testAPIKeySecret,
				envArticleTableName:     testArticlesTable,
				envUserProfileTableName: testProfilesTable,
				envSendsTableName:       testSendsTable,
				envCorsAllowOrigin:      tt.value,
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
		envStorageBackend: string(consts.StorageBackendSQLite),
		envSQLitePath:     testSQLitePath,
		envAPIKeySecret:   testAPIKeySecret,
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.StorageBackendSQLite, cfg.StorageBackend)
	assert.Equal(t, testSQLitePath, cfg.SQLitePath)
}

func TestLoad_DynamoDB_Backend_Success(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendDynamoDB),
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
	})

	awsLoader := mockAWSLoader()

	cfg, err := Load(consts.ModeServer, awsLoader)
	require.NoError(t, err)
	assert.Equal(t, consts.StorageBackendDynamoDB, cfg.StorageBackend)
	assert.NotNil(t, cfg.AWSConfig)
}

func TestLoad_Default_Storage_Backend(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envAPIKeySecret: testAPIKeySecret,
		envSQLitePath:   testSQLitePath,
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.StorageBackendSQLite, cfg.StorageBackend)
}

func TestLoad_SQLite_Path_Default(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend: string(consts.StorageBackendSQLite),
		envAPIKeySecret:   testAPIKeySecret,
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.StorageBackendSQLite, cfg.StorageBackend)
	assert.Equal(t, consts.SQLitePathDefault, cfg.SQLitePath)
}

func TestLoad_Missing_DynamoDB_Tables(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend: string(consts.StorageBackendDynamoDB),
		envAPIKeySecret:   testAPIKeySecret,
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	errMsg := err.Error()
	assert.Contains(t, errMsg, envArticleTableName)
	assert.Contains(t, errMsg, envUserProfileTableName)
	assert.Contains(t, errMsg, envSendsTableName)
}

func TestLoad_Invalid_Storage_Backend(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend: testInvalidBackend,
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported storage backend")
}

func TestLoad_Browserless_Key(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendSQLite),
		envSQLitePath:           testSQLitePath,
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
		envBrowserlessKey:       "browserless-key-123",
	})

	awsLoader := mockAWSLoader()

	cfg, err := Load(consts.ModeServer, awsLoader)
	require.NoError(t, err)
	assert.Equal(t, "browserless-key-123", cfg.BrowserlessKey)
}

func TestLoad_Process_Article_Lambda(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendSQLite),
		envSQLitePath:           testSQLitePath,
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
		envProcessArticleLambda: "arn:aws:lambda:us-east-1:123456789012:function:ProcessArticle",
	})

	awsLoader := mockAWSLoader()

	cfg, err := Load(consts.ModeServer, awsLoader)
	require.NoError(t, err)
	assert.Equal(t, "arn:aws:lambda:us-east-1:123456789012:function:ProcessArticle", cfg.ProcessArticleLambda)
}

func TestLoad_CLI_WithMailjet(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envEmailBackend:     string(consts.EmailBackendMailjet),
		envMailjetAPIKey:    testMailjetAPIKey,
		envMailjetAPISecret: testMailjetAPISecret,
		envSenderEmail:      testSenderEmail,
	})

	cfg, err := Load(consts.ModeCLI, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.EmailBackendMailjet, cfg.EmailProvider)
	assert.Equal(t, testMailjetAPIKey, cfg.MailjetAPIKey)
	assert.Equal(t, testMailjetAPISecret, cfg.MailjetAPISecret)
	assert.Equal(t, testSenderEmail, cfg.SenderEmail)
}

func TestLoad_InvalidEmailProvider(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envEmailBackend: "invalid-provider",
	})

	cfg, err := Load(consts.ModeCLI, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.EmailProvider("invalid-provider"), cfg.EmailProvider)
}

func TestLoad_MissingMailjetKeys_CLI(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envEmailBackend: string(consts.EmailBackendMailjet),
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
			expected: []string{testTask1},
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
				envStorageBackend:       string(consts.StorageBackendSQLite),
				envSQLitePath:           testSQLitePath,
				envAPIKeySecret:         testAPIKeySecret,
				envArticleTableName:     testArticlesTable,
				envUserProfileTableName: testProfilesTable,
				envSendsTableName:       testSendsTable,
			}
			if tt.tasksEnv != "" {
				envVars[envTasks] = tt.tasksEnv
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
		envStorageBackend:       string(consts.StorageBackendSQLite),
		envSQLitePath:           testSQLitePath,
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.DefaultHTTPPort, cfg.Port)
}

func TestLoad_CustomPort(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendSQLite),
		envSQLitePath:           testSQLitePath,
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
		envHTTPPort:             "3000",
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
		envHTTPPort: "9000",
	})

	cfg, err := Load(consts.ModeCLI, nil)
	require.NoError(t, err)
	assert.Equal(t, 9000, cfg.Port)
}

func TestLoad_CLI_InvalidStorageBackend(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend: testInvalidBackend,
		envAPIKeySecret:   testAPIKeySecret,
	})

	cfg, err := Load(consts.ModeCLI, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.StorageBackend(testInvalidBackend), cfg.StorageBackend)
}

func TestLoad_CLI_EmptyStorageBackend(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend: "",
		envAPIKeySecret:   testAPIKeySecret,
	})

	cfg, err := Load(consts.ModeCLI, nil)
	require.NoError(t, err)
	assert.Empty(t, cfg.StorageBackend)
}

func TestLoad_Server_SQLite_WithoutPath(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:   string(consts.StorageBackendSQLite),
		envAPIKeySecret:     testAPIKeySecret,
		envArticleTableName: testArticlesTable,
		testEnvProfileTable: testProfilesTable,
		envSendsTableName:   testSendsTable,
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	assert.Equal(t, consts.StorageBackendSQLite, cfg.StorageBackend)
	assert.Equal(t, consts.SQLitePathDefault, cfg.SQLitePath)
}

func TestLoad_Tasks_Config_InvalidJSON(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendSQLite),
		envSQLitePath:           testSQLitePath,
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
		envTasks:                "invalid-json{",
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	assert.Nil(t, cfg.Tasks)
}

func TestLoad_ArticleTagsTable(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendSQLite),
		envSQLitePath:           testSQLitePath,
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envArticleTagsTableName: "article-tags-table",
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	assert.Equal(t, "article-tags-table", cfg.ArticleTagsTable)
}

func TestLoad_BackupBucketName(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendSQLite),
		envSQLitePath:           testSQLitePath,
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
		envBackupBucketName:     "backup-bucket",
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
	assert.Contains(t, missing, envSQLitePath)
}

func TestValidateStorageBackendConfig_EmptyBackend(t *testing.T) {
	cfg := &Config{StorageBackend: "", SQLitePath: testSQLitePathAlt}
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
	assert.Contains(t, missing, envMailjetAPIKey)
	assert.Contains(t, missing, envMailjetAPISecret)
	assert.Contains(t, missing, envSenderEmail)
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
		SQLitePath:     testSQLitePathAlt,
		AuthBackend:    consts.AuthBackendSharedAPIKey,
		APIKeySecret:   testAPIKey,
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
		APIKeySecret:     testAPIKey,
	}
	var missing []string

	awsLoader := mockAWSLoader()

	err := cfg.validateServerConfig(&missing, awsLoader)
	require.NoError(t, err)
}

func TestLoad_Tasks_Config_InvalidTask(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendSQLite),
		envSQLitePath:           testSQLitePath,
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
		envTasks:                `[{"task":"` + testTask1 + `"}]`,
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	require.Len(t, cfg.Tasks, 1)
	assert.Equal(t, "task1", cfg.Tasks[0].Task)
	assert.Empty(t, cfg.Tasks[0].Schedule)
}

func TestLoad_Tasks_Config_WithBackup(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendSQLite),
		envSQLitePath:           testSQLitePath,
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
		envTasks:                `[{"task":"backup","schedule":"0 * * * *"}]`,
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	require.Len(t, cfg.Tasks, 1)
	assert.Equal(t, "backup", cfg.Tasks[0].Task)
	assert.Equal(t, "0 * * * *", cfg.Tasks[0].Schedule)
}

func TestLoad_Tasks_Config_WithRestore(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envStorageBackend:       string(consts.StorageBackendSQLite),
		envSQLitePath:           testSQLitePath,
		envAPIKeySecret:         testAPIKeySecret,
		envArticleTableName:     testArticlesTable,
		envUserProfileTableName: testProfilesTable,
		envSendsTableName:       testSendsTable,
		envTasks:                `[{"task":"restore","backup_name":"backup-123","overwrite":true}]`,
	})

	cfg, err := Load(consts.ModeServer, nil)
	require.NoError(t, err)
	require.Len(t, cfg.Tasks, 1)
	assert.Equal(t, "restore", cfg.Tasks[0].Task)
	assert.Equal(t, "backup-123", cfg.Tasks[0].BackupName)
	assert.True(t, cfg.Tasks[0].Overwrite)
}

func TestLoad_Auth0_MissingPASETOKey(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envAuthBackend:       testAuth0Str,
		envAuth0Domain:       testAuth0Domain,
		envAuth0Audience:     testAuth0Audience,
		envAuth0ClientID:     testAuth0ClientID,
		envAuth0ClientSecret: testAuth0ClientSecret,
		envPasetoKeyVersion:  testV1,
		envStorageBackend:    testSQLite,
		envSQLitePath:        testSQLitePath,
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), envPasetoKey)
}

func TestLoad_Auth0_MissingPASETOKeyVersion(t *testing.T) {
	setupEnvVars(t, map[string]string{
		envAuthBackend:       testAuth0Str,
		envAuth0Domain:       testAuth0Domain,
		envAuth0Audience:     testAuth0Audience,
		envAuth0ClientID:     testAuth0ClientID,
		envAuth0ClientSecret: testAuth0ClientSecret,
		envPasetoKey:         testPasetoKey,
		envStorageBackend:    testSQLite,
		envSQLitePath:        testSQLitePath,
	})

	awsLoader := mockAWSLoader()

	_, err := Load(consts.ModeServer, awsLoader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), envPasetoKeyVersion)
}
