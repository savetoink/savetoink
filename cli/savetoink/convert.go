package main

import (
	"context"
	"fmt"
	"log/slog"
	netURL "net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/model"
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
		return fmt.Errorf("missing email provider config: %s", strings.Join(missing, ", "))
	}

	return nil
}

func processArticle(ctx context.Context, url string, svc *service.Service) (*model.Article, []byte, error) {
	slog.Debug("processing article", slog.String("url", url))

	start := time.Now()

	u, err := netURL.Parse(url)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid url: %w", err)
	}

	fetched, err := svc.Fetch(ctx, u)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch article: %w", err)
	}

	doc, err := svc.ParseHTML(ctx, fetched)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse html: %w", err)
	}

	article, err := svc.Clean(ctx, doc, u)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to clean article: %w", err)
	}

	epubData, err := svc.GenerateEPUB(article)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate epub: %w", err)
	}

	slog.Debug("article processed", slog.Any("duration", time.Since(start)))

	return article, epubData, nil
}

func runConvert(_ *cobra.Command, args []string) error {
	url := args[0]

	if sendEmail {
		if validateErr := validateEmailConfig(cfg); validateErr != nil {
			return validateErr
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	svc := service.NewFromConfig(cfg)

	article, epubData, err := processArticle(ctx, url, svc)
	if err != nil {
		return err
	}

	if sendEmail {
		resp, emailErr := svc.SendArticle(ctx, destEmail, epubData, article.Title)
		if emailErr != nil {
			return fmt.Errorf("failed to send email: %w", emailErr)
		}
		slog.Info(fmt.Sprintf("✓ Article sent to e-reader device at %s (email ID: %s)", destEmail, resp.MessageID))
	} else {
		if outputPath == "" {
			outputPath = "article.epub"
		}
		if writeErr := os.WriteFile(outputPath, epubData, defaultFilePerms); writeErr != nil {
			return fmt.Errorf("failed to write EPUB: %w", writeErr)
		}
		absPath, _ := filepath.Abs(outputPath)
		slog.Info("✓ EPUB saved to " + absPath)
	}

	return nil
}
