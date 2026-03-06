// Package config provides configuration for the savetoink application.
package config

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/shaftoe/savetoink/backend/consts"
	"github.com/spf13/viper"
)

// Config holds configuration settings for application.
type Config struct {
	Debug            bool
	ArticlesTable    string
	UserProfileTable string
	SendsTable       string
	AppURL           string
	Mode             consts.RunMode

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
}

// Load reads configuration from environment variables and returns a Config instance.
func Load(mode consts.RunMode) (*Config, error) {
	viper.SetEnvPrefix("SAVETOINK")
	viper.AutomaticEnv()

	if err := bindEnvVars(); err != nil {
		return nil, err
	}

	cfg := loadConfig(mode)

	if cfg.AuthBackend == "" {
		cfg.AuthBackend = consts.AuthBackendSharedAPIKey
	}

	if err := cfg.validate(); err != nil {
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
		{"app-url", "SAVETOINK_APP_URL"},
		{"articles-table", "SAVETOINK_ARTICLE_TABLE_NAME"},
		{"auth-backend", "SAVETOINK_AUTH_BACKEND"},
		{"auth0-audience", "SAVETOINK_AUTH0_AUDIENCE"},
		{"auth0-client-id", "SAVETOINK_AUTH0_CLIENT_ID"},
		{"auth0-client-secret", "SAVETOINK_AUTH0_CLIENT_SECRET"},
		{"auth0-domain", "SAVETOINK_AUTH0_DOMAIN"},
		{"debug", "SAVETOINK_DEBUG"},
		{"email-backend", "SAVETOINK_EMAIL_BACKEND"},
		{"sender-email", "SAVETOINK_SENDER_EMAIL"},
		{"sends-table", "SAVETOINK_SENDS_TABLE_NAME"},
		{"user-profile-table", "SAVETOINK_USER_PROFILE_TABLE_NAME"},
		{"logging-provider", "SAVETOINK_LOGGING_PROVIDER"},
		{"sentry-dsn", "SAVETOINK_SENTRY_DSN"},
		{"sentry-environment", "SAVETOINK_SENTRY_ENVIRONMENT"},
		{"sentry-sample-rate", "SAVETOINK_SENTRY_SAMPLE_RATE"},
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
		AppURL:               viper.GetString("app-url"),
		ArticlesTable:        viper.GetString("articles-table"),
		Auth0Audience:        viper.GetString("auth0-audience"),
		Auth0ClientID:        viper.GetString("auth0-client-id"),
		Auth0ClientSecret:    viper.GetString("auth0-client-secret"),
		Auth0Domain:          viper.GetString("auth0-domain"),
		AuthBackend:          consts.AuthBackend(viper.GetString("auth-backend")),
		Debug:                viper.GetBool("debug"),
		EmailProvider:        consts.EmailProvider(viper.GetString("email-backend")),
		MailjetAPIKey:        viper.GetString("api-key"),
		MailjetAPISecret:     viper.GetString("api-secret"),
		MailjetWebhookSecret: viper.GetString("api-webhook-secret"),
		Mode:                 mode,
		SenderEmail:          viper.GetString("sender-email"),
		SendsTable:           viper.GetString("sends-table"),
		UserProfileTable:     viper.GetString("user-profile-table"),
		LoggingProvider:      consts.LoggingProvider(viper.GetString("logging-provider")),
		SentryDSN:            viper.GetString("sentry-dsn"),
		SentryEnvironment:    viper.GetString("sentry-environment"),
		SentrySampleRate:     viper.GetFloat64("sentry-sample-rate"),
	}

	return cfg
}

func (c *Config) validate() error {
	var missing []string

	if c.Mode == consts.ModeServer {
		if err := c.validateServerConfig(&missing); err != nil {
			return err
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("required environment variables are missing: %v", missing)
	}

	return nil
}

func (c *Config) validateServerConfig(missing *[]string) error {
	if err := c.validateAuthBackendConfig(missing); err != nil {
		return err
	}

	c.validateEmailProviderConfig(missing)
	c.validateLoggingProviderConfig(missing)

	if c.ArticlesTable == "" {
		*missing = append(*missing, "SAVETOINK_ARTICLE_TABLE_NAME")
	}
	if c.UserProfileTable == "" {
		*missing = append(*missing, "SAVETOINK_USER_PROFILE_TABLE_NAME")
	}
	if c.SendsTable == "" {
		*missing = append(*missing, "SAVETOINK_SENDS_TABLE_NAME")
	}
	if c.AppURL == "" {
		*missing = append(*missing, "SAVETOINK_APP_URL")
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}
	c.AWSConfig = &cfg

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

func (c *Config) validateEmailProviderConfig(missing *[]string) {
	if c.EmailProvider == consts.EmailBackendMailjet {
		if c.MailjetAPIKey == "" {
			*missing = append(*missing, "SAVETOINK_MAILJET_API_KEY")
		}
		if c.MailjetAPISecret == "" {
			*missing = append(*missing, "SAVETOINK_MAILJET_API_SECRET")
		}
		if c.MailjetWebhookSecret == "" {
			*missing = append(*missing, "SAVETOINK_MAILJET_WEBHOOK_SECRET")
		}
		if c.SenderEmail == "" {
			*missing = append(*missing, "SAVETOINK_SENDER_EMAIL")
		}
	}
}

func (c *Config) validateLoggingProviderConfig(missing *[]string) {
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
}
