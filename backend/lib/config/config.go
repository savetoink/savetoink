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
		{"api-key", "SAVETOINK_MAILJET_API_KEY"},
		{"api-key-secret", "SAVETOINK_API_KEY"},
		{"api-secret", "SAVETOINK_MAILJET_API_SECRET"},
		{"api-webhook-secret", "SAVETOINK_MAILJET_WEBHOOK_SECRET"},
		{"article-tags-table", "SAVETOINK_ARTICLE_TAGS_TABLE_NAME"},
		{"articles-table", "SAVETOINK_ARTICLE_TABLE_NAME"},
		{"auth-backend", "SAVETOINK_AUTH_BACKEND"},
		{"backup-bucket-name", "SAVETOINK_BACKUP_BUCKET_NAME"},
		{"auth0-audience", "SAVETOINK_AUTH0_AUDIENCE"},
		{"auth0-client-id", "SAVETOINK_AUTH0_CLIENT_ID"},
		{"auth0-client-secret", "SAVETOINK_AUTH0_CLIENT_SECRET"},
		{"auth0-domain", "SAVETOINK_AUTH0_DOMAIN"},
		{"browserless-key", "SAVETOINK_BROWSERLESS_KEY"},
		{"cors-allow-origin", "SAVETOINK_CORS_ALLOW_ORIGIN"},
		{"debug", "SAVETOINK_DEBUG"},
		{"email-backend", "SAVETOINK_EMAIL_BACKEND"},
		{"http-port", "SAVETOINK_HTTP_PORT"},
		{"logging-provider", "SAVETOINK_LOGGING_PROVIDER"},
		{"process-article-lambda", "SAVETOINK_PROCESS_ARTICLE_LAMBDA"},
		{"sender-email", "SAVETOINK_SENDER_EMAIL"},
		{"sends-table", "SAVETOINK_SENDS_TABLE_NAME"},
		{"sentry-dsn", "SAVETOINK_SENTRY_DSN"},
		{"sentry-environment", "SAVETOINK_SENTRY_ENVIRONMENT"},
		{"sentry-sample-rate", "SAVETOINK_SENTRY_SAMPLE_RATE"},
		{"sqlite-path", "SAVETOINK_SQLITE_PATH"},
		{"storage-backend", "SAVETOINK_STORAGE_BACKEND"},
		{"tasks", "SAVETOINK_TASKS"},
		{"user-profile-table", "SAVETOINK_USER_PROFILE_TABLE_NAME"},
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
		APIKeySecret:         viper.GetString("api-key-secret"),
		ArticleTagsTable:     viper.GetString("article-tags-table"),
		ArticlesTable:        viper.GetString("articles-table"),
		Auth0Audience:        viper.GetString("auth0-audience"),
		Auth0ClientID:        viper.GetString("auth0-client-id"),
		Auth0ClientSecret:    viper.GetString("auth0-client-secret"),
		Auth0Domain:          viper.GetString("auth0-domain"),
		AuthBackend:          consts.AuthBackend(viper.GetString("auth-backend")),
		BackupBucketName:     viper.GetString("backup-bucket-name"),
		BrowserlessKey:       viper.GetString("browserless-key"),
		CorsAllowOrigin:      viper.GetString("cors-allow-origin"),
		Debug:                viper.GetBool("debug"),
		EmailProvider:        consts.EmailProvider(viper.GetString("email-backend")),
		LoggingProvider:      consts.LoggingProvider(viper.GetString("logging-provider")),
		MailjetAPIKey:        viper.GetString("api-key"),
		MailjetAPISecret:     viper.GetString("api-secret"),
		Port:                 viper.GetInt("http-port"),
		MailjetWebhookSecret: viper.GetString("api-webhook-secret"),
		Mode:                 mode,
		ProcessArticleLambda: viper.GetString("process-article-lambda"),
		SQLitePath:           viper.GetString("sqlite-path"),
		SenderEmail:          viper.GetString("sender-email"),
		SendsTable:           viper.GetString("sends-table"),
		SentryDSN:            viper.GetString("sentry-dsn"),
		SentryEnvironment:    viper.GetString("sentry-environment"),
		SentrySampleRate:     viper.GetFloat64("sentry-sample-rate"),
		StorageBackend:       consts.StorageBackend(viper.GetString("storage-backend")),
		Tasks:                loadTasksConfig(),
		UserProfileTable:     viper.GetString("user-profile-table"),
	}

	return cfg
}

func loadTasksConfig() []consts.TaskConfig {
	tasksJSON := viper.GetString("tasks")
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
		*missing = append(*missing, "SAVETOINK_ARTICLE_TABLE_NAME")
	}
	if c.UserProfileTable == "" {
		*missing = append(*missing, "SAVETOINK_USER_PROFILE_TABLE_NAME")
	}
	if c.SendsTable == "" {
		*missing = append(*missing, "SAVETOINK_SENDS_TABLE_NAME")
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
		*missing = append(*missing, "SAVETOINK_SQLITE_PATH")
	}
	return nil
}

func (c *Config) validateAuthBackendConfig(missing *[]string) error {
	switch c.AuthBackend {
	case consts.AuthBackendSharedAPIKey:
		if c.APIKeySecret == "" {
			*missing = append(*missing, "SAVETOINK_API_KEY")
		}
	case consts.AuthBackendAuth0:
		if c.Auth0Domain == "" {
			*missing = append(*missing, "SAVETOINK_AUTH0_DOMAIN")
		}
		if c.Auth0Audience == "" {
			*missing = append(*missing, "SAVETOINK_AUTH0_AUDIENCE")
		}
		if c.Auth0ClientID == "" {
			*missing = append(*missing, "SAVETOINK_AUTH0_CLIENT_ID")
		}
		if c.Auth0ClientSecret == "" {
			*missing = append(*missing, "SAVETOINK_AUTH0_CLIENT_SECRET")
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
			*missing = append(*missing, "SAVETOINK_MAILJET_WEBHOOK_SECRET")
		}
	}
	return nil
}

func (c *Config) validateLoggingProviderConfig(missing *[]string) error {
	if c.LoggingProvider == consts.LoggingBackendSentry {
		if c.SentryDSN == "" {
			*missing = append(*missing, "SAVETOINK_SENTRY_DSN")
		}
		if c.SentryEnvironment == "" {
			*missing = append(*missing, "SAVETOINK_SENTRY_ENVIRONMENT")
		}
		if c.SentrySampleRate == 0 {
			*missing = append(*missing, "SAVETOINK_SENTRY_SAMPLE_RATE")
		}
	}
	return nil
}

// ValidateEmailProviderConfigCli validates of email provider config for CLI.
func (c *Config) ValidateEmailProviderConfigCli(missing *[]string) {
	if c.EmailProvider == consts.EmailBackendMailjet {
		if c.MailjetAPIKey == "" {
			*missing = append(*missing, "SAVETOINK_MAILJET_API_KEY")
		}
		if c.MailjetAPISecret == "" {
			*missing = append(*missing, "SAVETOINK_MAILJET_API_SECRET")
		}
		if c.SenderEmail == "" {
			*missing = append(*missing, "SAVETOINK_SENDER_EMAIL")
		}
	}
}
