package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/shaftoe/savetoink/backend/lib/service/content"
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

	fetcher := content.NewFetcher("")
	extractor := content.NewDOMExtractor()
	cleaner := content.NewTrafilaturaCleaner()

	fetched, err := fetcher.Fetch(ctx, u)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer func() {
		if closeErr := fetched.HTML.Close(); closeErr != nil {
			slog.Warn("failed to close response", slog.Any("error", closeErr))
		}
	}()

	doc, err := extractor.Extract(ctx, fetched.HTML)
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	article, err := cleaner.Clean(ctx, doc, u)
	if err != nil {
		return fmt.Errorf("failed to clean content: %w", err)
	}

	fmt.Print(article.Content)
	return nil
}
