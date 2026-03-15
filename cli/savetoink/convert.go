package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	netURL "net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/service"
	"github.com/spf13/cobra"
)

func isWebURL(input string) bool {
	_, err := netURL.ParseRequestURI(input)
	return err == nil && (hasScheme(input) || strings.HasPrefix(input, "localhost"))
}

func hasScheme(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func processArticle(ctx context.Context, input string, svc *service.Service) (io.ReadCloser, *string, error) {
	slog.Debug("processing article", slog.String("input", input))

	start := time.Now()

	var u *netURL.URL
	var err error

	if isWebURL(input) {
		u, err = netURL.Parse(input)
	} else {
		absPath, absErr := filepath.Abs(input)
		if absErr != nil {
			return nil, nil, fmt.Errorf("failed to get absolute path: %w", absErr)
		}
		u = &netURL.URL{
			Scheme: "file",
			Path:   absPath,
		}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("invalid input: %w", err)
	}

	doc, err := svc.ParseHTMLFromSource(ctx, u)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse html: %w", err)
	}

	article, err := svc.Clean(ctx, doc, u)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to clean article: %w", err)
	}

	epubReader, err := svc.GenerateEPUB(article)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate epub: %w", err)
	}

	slog.Debug("article processed", slog.Any("duration", time.Since(start)))

	return epubReader, &article.Title, nil
}

func runConvert(_ *cobra.Command, args []string) error {
	input := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	svc := service.NewFromConfig(cfg)

	epubReader, _, err := processArticle(ctx, input, svc)
	if err != nil {
		return err
	}
	defer func() { _ = epubReader.Close() }()

	epubData, err := io.ReadAll(epubReader)
	if err != nil {
		return fmt.Errorf("failed to read epub data: %w", err)
	}

	output := outputPath
	if output == "" {
		fmt.Print(string(epubData))
	} else {
		if writeErr := os.WriteFile(output, epubData, defaultFilePerms); writeErr != nil {
			return fmt.Errorf("failed to write EPUB: %w", writeErr)
		}
		absPath, _ := filepath.Abs(output)
		slog.Info("✓ EPUB saved to " + absPath)
	}

	return nil
}
