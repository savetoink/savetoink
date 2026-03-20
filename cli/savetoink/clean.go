package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/shaftoe/savetoink/backend/lib/service"
	"github.com/spf13/cobra"
)

func runClean(cmd *cobra.Command, args []string) error {
	inputURL := args[0]

	u, err := url.Parse(inputURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
	defer cancel()

	svc := service.NewFromConfig(cfg)

	doc, err := svc.ParseHTMLFromSource(ctx, u)
	if err != nil {
		return fmt.Errorf("failed to parse html: %w", err)
	}

	article, err := svc.Clean(ctx, doc, u)
	if err != nil {
		return fmt.Errorf("failed to clean content: %w", err)
	}

	output := outputPath
	if output == "" {
		fmt.Print(article.Content)
	} else {
		if writeErr := os.WriteFile(output, []byte(article.Content), defaultFilePerms); writeErr != nil {
			return fmt.Errorf("failed to write output: %w", writeErr)
		}
		slog.Info("✓ cleaned content saved to " + output)
	}
	return nil
}
