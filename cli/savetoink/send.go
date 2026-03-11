package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/service"
	"github.com/spf13/cobra"
)

func validateEmailConfig(cfg *config.Config) error {
	if cfg.EmailProvider != consts.EmailBackendMailjet {
		return fmt.Errorf("missing or unsupported email provider: '%s'", cfg.EmailProvider)
	}

	var missing []string
	cfg.ValidateEmailProviderConfigCli(&missing)

	if len(missing) > 0 {
		return fmt.Errorf("missing email provider config: %s", missing)
	}

	return nil
}

func runSend(_ *cobra.Command, args []string) error {
	input := args[0]

	if destEmail == "" {
		return errors.New("missing required flag: --dest-email")
	}

	validateErr := validateEmailConfig(cfg)
	if validateErr != nil {
		return validateErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	svc := service.NewFromConfig(cfg)

	epubData, err := processArticle(ctx, input, svc)
	if err != nil {
		return err
	}

	resp, emailErr := svc.SendArticle(ctx, destEmail, epubData, "")
	if emailErr != nil {
		return fmt.Errorf("failed to send email: %w", emailErr)
	}
	slog.Info(fmt.Sprintf("✓ Article sent to e-reader device at %s (email ID: %s)", destEmail, resp.MessageID))

	return nil
}
