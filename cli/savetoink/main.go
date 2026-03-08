// Package main implements the CLI application for converting web articles to EPUB.
package main

import (
	"context"
	"fmt"
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

const (
	defaultTimeoutSeconds = 30
	defaultFilePerms      = 0o600
)

var (
	outputPath string
	timeout    time.Duration

	sendEmail bool
	destEmail string
)

var rootCmd = &cobra.Command{
	Use:   "savetoink",
	Short: "Convert web articles to EPUB format",
	Long:  `A CLI tool to fetch web articles and convert them to EPUB format for Kindle devices.`,
}

var convertCmd = &cobra.Command{
	Use:   "convert [url]",
	Short: "Convert a URL to EPUB",
	Long: `Fetch a web article from given URL and convert it to EPUB format.
 Use --send to skip local EPUB generation and send converted EPUB to your Kindle.`,
	Args: cobra.ExactArgs(1),
	RunE: runConvert,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number of the savetoink application.`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(*consts.Version())
	},
}

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
	fmt.Printf("Processing article from: %s\n", url)

	start := time.Now()

	htmlBytes, _, err := svc.Fetch(ctx, url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch article: %w", err)
	}

	article, err := svc.Extract(ctx, htmlBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract article: %w", err)
	}

	fmt.Printf("Processed in %v\n", time.Since(start))

	epubData, err := svc.GenerateEPUB(article)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate epub: %w", err)
	}

	return article, epubData, nil
}

func runConvert(_ *cobra.Command, args []string) error {
	url := args[0]

	cfg, err := config.Load(consts.ModeCLI, nil)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

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
		fmt.Printf("\n✓ Article sent to e-reader device (email ID: %s)\n", resp.MessageID)
	} else {
		if outputPath == "" {
			outputPath = "article.epub"
		}
		if writeErr := os.WriteFile(outputPath, epubData, defaultFilePerms); writeErr != nil {
			return fmt.Errorf("failed to write EPUB: %w", writeErr)
		}
		absPath, _ := filepath.Abs(outputPath)
		fmt.Printf("\n✓ EPUB saved to: %s\n", absPath)
	}

	return nil
}

func main() {
	convertCmd.Flags().StringVarP(&outputPath, "output", "o", "article.epub", "Output file path")
	convertCmd.Flags().DurationVarP(&timeout, "timeout", "t",
		defaultTimeoutSeconds*time.Second, "Timeout for HTTP requests")

	convertCmd.Flags().BoolVar(&sendEmail, "send", false, "Send EPUB to Kindle via email instead of saving locally")
	convertCmd.Flags().StringVar(&destEmail, "dest-email", "", "Destination Kindle email address")

	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
