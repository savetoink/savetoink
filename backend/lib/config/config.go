// Package config provides configuration for the savetoink application.
package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/spf13/viper"
)

// Environment variable names.
const (
	envAPIKeySecret         = "SAVETOINK_API_KEY" //nolint:gosec // env var name, not a credential
	envArticleTableName     = "SAVETOINK_ARTICLE_TABLE_NAME"
	envArticleTagsTableName = "SAVETOINK_ARTICLE_TAGS_TABLE_NAME"
	envAuth0Audience        = "SAVETOINK_AUTH0_AUDIENCE"
	envAuth0ClientID        = "SAVETOINK_AUTH0_CLIENT_ID"
	envAuth0ClientSecret    = "SAVETOINK_AUTH0_CLIENT_SECRET" //nolint:gosec // env var name, not a credential
	envAuth0Domain          = "SAVETOINK_AUTH0_DOMAIN"
	envAuthBackend          = "SAVETOINK_AUTH_BACKEND"
	envBackupBucketName     = "SAVETOINK_BACKUP_BUCKET_NAME"
	envBrowserlessKey       = "SAVETOINK_BROWSERLESS_KEY"
	envCorsAllowOrigin      = "SAVETOINK_CORS_ALLOW_ORIGIN"
	envDebug                = "SAVETOINK_DEBUG"
	envDisableQuotaCheck    = "SAVETOINK_DISABLE_QUOTA_CHECK"
	envEmailBackend         = "SAVETOINK_EMAIL_BACKEND"
	envHTTPPort             = "SAVETOINK_HTTP_PORT"
	envLoggingProvider      = "SAVETOINK_LOGGING_PROVIDER"
	envMailjetAPIKey        = "SAVETOINK_MAILJET_API_KEY"        //nolint:gosec // env var name, not a credential
	envMailjetAPISecret     = "SAVETOINK_MAILJET_API_SECRET"     //nolint:gosec // env var name, not a credential
	envMailjetWebhookSecret = "SAVETOINK_MAILJET_WEBHOOK_SECRET" //nolint:gosec // env var name, not a credential
	envPasetoKey            = "SAVETOINK_PASETO_KEY"
	envPasetoKeyVersion     = "SAVETOINK_PASETO_KEY_VERSION"
	envProcessArticleLambda = "SAVETOINK_PROCESS_ARTICLE_LAMBDA"
	envSenderEmail          = "SAVETOINK_SENDER_EMAIL"
	envSendsTableName       = "SAVETOINK_SENDS_TABLE_NAME"
	envSentryDSN            = "SAVETOINK_SENTRY_DSN"
	envSentryEnvironment    = "SAVETOINK_SENTRY_ENVIRONMENT"
	envSentrySampleRate     = "SAVETOINK_SENTRY_SAMPLE_RATE"
	envSQLitePath           = "SAVETOINK_SQLITE_PATH"
	envStorageBackend       = "SAVETOINK_STORAGE_BACKEND"
	envTasks                = "SAVETOINK_TASKS"
	envUserProfileTableName = "SAVETOINK_USER_PROFILE_TABLE_NAME"
)

// Viper config keys.
const (
	viperAPIKey               = "api-key"
	viperAPIKeySecret         = "api-key-secret"
	viperAPISecret            = "api-secret"
	viperAPIWebhookSecret     = "api-webhook-secret" //nolint:gosec // viper key name, not a credential
	viperArticleTagsTable     = "article-tags-table"
	viperArticlesTable        = "articles-table"
	viperAuth0Audience        = "auth0-audience"
	viperAuth0ClientID        = "auth0-client-id"
	viperAuth0ClientSecret    = "auth0-client-secret" //nolint:gosec // viper key name, not a credential
	viperAuth0Domain          = "auth0-domain"
	viperAuthBackend          = "auth-backend"
	viperBackupBucketName     = "backup-bucket-name"
	viperBrowserlessKey       = "browserless-key"
	viperCorsAllowOrigin      = "cors-allow-origin"
	viperDebug                = "debug"
	viperDisableQuotaCheck    = "disable-quota-check"
	viperEmailBackend         = "email-backend"
	viperHTTPPort             = "http-port"
	viperLoggingProvider      = "logging-provider"
	viperPasetoKey            = "paseto-key"
	viperPasetoKeyVersion     = "paseto-key-version"
	viperProcessArticleLambda = "process-article-lambda"
	viperSenderEmail          = "sender-email"
	viperSendsTable           = "sends-table"
	viperSentryDSN            = "sentry-dsn"
	viperSentryEnvironment    = "sentry-environment"
	viperSentrySampleRate     = "sentry-sample-rate"
	viperSQLitePath           = "sqlite-path"
	viperStorageBackend       = "storage-backend"
	viperTasks                = "tasks"
	viperUserProfileTable     = "user-profile-table"
)

// Config holds configuration settings for application.
type Config struct {
	Debug            bool
	ArticlesTable    string
	ArticleTagsTable string
	UserProfileTable string
	SendsTable       string
	BackupBucketName string
	Mode             consts.RunMode
	CorsAllowOrigin  string
	Port             int

	// Storage
	StorageBackend consts.StorageBackend
	SQLitePath     string

	// Auth
	APIKeySecret      string
	AWSConfig         *aws.Config
	Auth0Audience     string
	Auth0ClientID     string
	Auth0ClientSecret string
	Auth0Domain       string
	AuthBackend       consts.AuthBackend

	// PASETO
	PASETOSymmetricKey string
	PASETOKeyVersion   string

	// Email
	EmailProvider        consts.EmailProvider
	MailjetAPIKey        string
	MailjetAPISecret     string
	MailjetWebhookSecret string
	SenderEmail          string

	// Logging
	LoggingProvider   consts.LoggingProvider
	SentryDSN         string
	SentryEnvironment string
	SentrySampleRate  float64

	// Content Fetching
	BrowserlessKey string

	// Lambda
	ProcessArticleLambda string

	// Tasks
	Tasks []consts.TaskConfig

	// Quota
	DisableQuotaCheck bool
}

// AWSConfigLoader is a function type for loading AWS configuration.
// This allows for dependency injection and easier testing.
type AWSConfigLoader func(ctx context.Context) (aws.Config, error)

// Load reads configuration from environment variables and returns a Config instance.
// The awsLoader parameter is used to load AWS configuration in server mode.
// It can be nil for CLI mode or in tests where AWS config is not needed.
func Load(mode consts.RunMode, awsLoader AWSConfigLoader) (*Config, error) {
	viper.SetEnvPrefix("SAVETOINK")
	viper.AutomaticEnv()

	if err := bindEnvVars(); err != nil {
		return nil, err
	}

	cfg := loadConfig(mode)

	if cfg.AuthBackend == "" {
		cfg.AuthBackend = consts.AuthBackendSharedAPIKey
	}

	if cfg.Port == 0 {
		cfg.Port = consts.DefaultHTTPPort
	}

	if cfg.CorsAllowOrigin == "" {
		cfg.CorsAllowOrigin = "*"
	}

	if err := cfg.validate(awsLoader); err != nil {
		return nil, err
	}

	return cfg, nil
}

func bindEnvVars() error {
	envVars := []struct {
		key    string
		envVar string
	}{
		{viperAPIKey, envMailjetAPIKey},
		{viperAPIKeySecret, envAPIKeySecret},
		{viperAPISecret, envMailjetAPISecret},
		{viperAPIWebhookSecret, envMailjetWebhookSecret},
		{viperArticleTagsTable, envArticleTagsTableName},
		{viperArticlesTable, envArticleTableName},
		{viperAuthBackend, envAuthBackend},
		{viperBackupBucketName, envBackupBucketName},
		{viperAuth0Audience, envAuth0Audience},
		{viperAuth0ClientID, envAuth0ClientID},
		{viperAuth0ClientSecret, envAuth0ClientSecret},
		{viperAuth0Domain, envAuth0Domain},
		{viperBrowserlessKey, envBrowserlessKey},
		{viperCorsAllowOrigin, envCorsAllowOrigin},
		{viperDebug, envDebug},
		{viperDisableQuotaCheck, envDisableQuotaCheck},
		{viperEmailBackend, envEmailBackend},
		{viperHTTPPort, envHTTPPort},
		{viperLoggingProvider, envLoggingProvider},
		{viperPasetoKey, envPasetoKey},
		{viperPasetoKeyVersion, envPasetoKeyVersion},
		{viperProcessArticleLambda, envProcessArticleLambda},
		{viperSenderEmail, envSenderEmail},
		{viperSendsTable, envSendsTableName},
		{viperSentryDSN, envSentryDSN},
		{viperSentryEnvironment, envSentryEnvironment},
		{viperSentrySampleRate, envSentrySampleRate},
		{viperSQLitePath, envSQLitePath},
		{viperStorageBackend, envStorageBackend},
		{viperTasks, envTasks},
		{viperUserProfileTable, envUserProfileTableName},
	}

	for _, ev := range envVars {
		if err := viper.BindEnv(ev.key, ev.envVar); err != nil {
			return fmt.Errorf("failed to bind %s env: %w", ev.key, err)
		}
	}
	return nil
}

func loadConfig(mode consts.RunMode) *Config {
	cfg := &Config{
		APIKeySecret:         viper.GetString(viperAPIKeySecret),
		ArticleTagsTable:     viper.GetString(viperArticleTagsTable),
		ArticlesTable:        viper.GetString(viperArticlesTable),
		Auth0Audience:        viper.GetString(viperAuth0Audience),
		Auth0ClientID:        viper.GetString(viperAuth0ClientID),
		Auth0ClientSecret:    viper.GetString(viperAuth0ClientSecret),
		Auth0Domain:          viper.GetString(viperAuth0Domain),
		AuthBackend:          consts.AuthBackend(viper.GetString(viperAuthBackend)),
		BackupBucketName:     viper.GetString(viperBackupBucketName),
		BrowserlessKey:       viper.GetString(viperBrowserlessKey),
		CorsAllowOrigin:      viper.GetString(viperCorsAllowOrigin),
		Debug:                viper.GetBool(viperDebug),
		DisableQuotaCheck:    viper.GetBool(viperDisableQuotaCheck),
		EmailProvider:        consts.EmailProvider(viper.GetString(viperEmailBackend)),
		LoggingProvider:      consts.LoggingProvider(viper.GetString(viperLoggingProvider)),
		MailjetAPIKey:        viper.GetString(viperAPIKey),
		MailjetAPISecret:     viper.GetString(viperAPISecret),
		Port:                 viper.GetInt(viperHTTPPort),
		MailjetWebhookSecret: viper.GetString(viperAPIWebhookSecret),
		Mode:                 mode,
		PASETOSymmetricKey:   viper.GetString(viperPasetoKey),
		PASETOKeyVersion:     viper.GetString(viperPasetoKeyVersion),
		ProcessArticleLambda: viper.GetString(viperProcessArticleLambda),
		SQLitePath:           viper.GetString(viperSQLitePath),
		SenderEmail:          viper.GetString(viperSenderEmail),
		SendsTable:           viper.GetString(viperSendsTable),
		SentryDSN:            viper.GetString(viperSentryDSN),
		SentryEnvironment:    viper.GetString(viperSentryEnvironment),
		SentrySampleRate:     viper.GetFloat64(viperSentrySampleRate),
		StorageBackend:       consts.StorageBackend(viper.GetString(viperStorageBackend)),
		Tasks:                loadTasksConfig(),
		UserProfileTable:     viper.GetString(viperUserProfileTable),
	}

	return cfg
}

func loadTasksConfig() []consts.TaskConfig {
	tasksJSON := viper.GetString(viperTasks)
	if tasksJSON == "" {
		return nil
	}

	var tasks []consts.TaskConfig
	if err := json.Unmarshal([]byte(tasksJSON), &tasks); err != nil {
		slog.With("version", *consts.Version()).
			Error("failed to unmarshal tasks config, background scheduler will not be loaded",
				slog.Any("error", err))
		return nil
	}

	return tasks
}

func (c *Config) validate(awsLoader AWSConfigLoader) error {
	var missing []string

	if c.Mode == consts.ModeServer {
		if err := c.validateServerConfig(&missing, awsLoader); err != nil {
			return err
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("required environment variables are missing: %v", missing)
	}

	return nil
}

func (c *Config) validateServerConfig(missing *[]string, awsLoader AWSConfigLoader) error {
	if err := c.validateAuthBackendConfig(missing); err != nil {
		return err
	}

	if err := c.validateEmailProviderConfig(missing); err != nil {
		return err
	}

	if err := c.validateLoggingProviderConfig(missing); err != nil {
		return err
	}

	return c.validateStorageBackendConfig(missing, awsLoader)
}

func (c *Config) validateStorageBackendConfig(missing *[]string, awsLoader AWSConfigLoader) error {
	if c.StorageBackend == "" {
		c.StorageBackend = consts.StorageBackendSQLite
	}

	switch c.StorageBackend {
	case consts.StorageBackendDynamoDB:
		if err := c.validateDynamoDBConfig(missing, awsLoader); err != nil {
			return err
		}
	case consts.StorageBackendSQLite:
		if c.SQLitePath == "" {
			c.SQLitePath = consts.SQLitePathDefault
		}
		if err := c.validateSQLiteConfig(missing); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported storage backend: %s", c.StorageBackend)
	}

	return nil
}

func (c *Config) validateDynamoDBConfig(missing *[]string, awsLoader AWSConfigLoader) error {
	if c.ArticlesTable == "" {
		*missing = append(*missing, envArticleTableName)
	}
	if c.UserProfileTable == "" {
		*missing = append(*missing, envUserProfileTableName)
	}
	if c.SendsTable == "" {
		*missing = append(*missing, envSendsTableName)
	}
	return c.validateAWSLoaderConfig(awsLoader)
}

func (c *Config) validateAWSLoaderConfig(awsLoader AWSConfigLoader) error {
	if awsLoader == nil {
		return errors.New("AWS config loader is required in server mode")
	}
	cfg, err := awsLoader(context.Background())
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}
	c.AWSConfig = &cfg
	return nil
}

func (c *Config) validateSQLiteConfig(missing *[]string) error {
	if c.SQLitePath == "" {
		*missing = append(*missing, envSQLitePath)
	}
	return nil
}

func (c *Config) validateAuthBackendConfig(missing *[]string) error {
	switch c.AuthBackend {
	case consts.AuthBackendSharedAPIKey:
		if c.APIKeySecret == "" {
			*missing = append(*missing, envAPIKeySecret)
		}
	case consts.AuthBackendAuth0:
		if c.Auth0Domain == "" {
			*missing = append(*missing, envAuth0Domain)
		}
		if c.Auth0Audience == "" {
			*missing = append(*missing, envAuth0Audience)
		}
		if c.Auth0ClientID == "" {
			*missing = append(*missing, envAuth0ClientID)
		}
		if c.Auth0ClientSecret == "" {
			*missing = append(*missing, envAuth0ClientSecret)
		}
		if c.PASETOSymmetricKey == "" {
			*missing = append(*missing, envPasetoKey)
		}
		if c.PASETOKeyVersion == "" {
			*missing = append(*missing, envPasetoKeyVersion)
		}
	default:
		return fmt.Errorf("unsupported auth backend: %s", c.AuthBackend)
	}
	return nil
}

func (c *Config) validateEmailProviderConfig(missing *[]string) error {
	c.ValidateEmailProviderConfigCli(missing)
	if c.EmailProvider == consts.EmailBackendMailjet {
		if c.MailjetWebhookSecret == "" {
			*missing = append(*missing, envMailjetWebhookSecret)
		}
	}
	return nil
}

func (c *Config) validateLoggingProviderConfig(missing *[]string) error {
	if c.LoggingProvider == consts.LoggingBackendSentry {
		if c.SentryDSN == "" {
			*missing = append(*missing, envSentryDSN)
		}
		if c.SentryEnvironment == "" {
			*missing = append(*missing, envSentryEnvironment)
		}
		if c.SentrySampleRate == 0 {
			*missing = append(*missing, envSentrySampleRate)
		}
	}
	return nil
}

// ValidateEmailProviderConfigCli validates of email provider config for CLI.
func (c *Config) ValidateEmailProviderConfigCli(missing *[]string) {
	if c.EmailProvider == consts.EmailBackendMailjet {
		if c.MailjetAPIKey == "" {
			*missing = append(*missing, envMailjetAPIKey)
		}
		if c.MailjetAPISecret == "" {
			*missing = append(*missing, envMailjetAPISecret)
		}
		if c.SenderEmail == "" {
			*missing = append(*missing, envSenderEmail)
		}
	}
}
