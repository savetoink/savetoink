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
	"github.com/shaftoe/savetoink/backend/lib/email"
	"github.com/shaftoe/savetoink/backend/lib/service"
	"github.com/shaftoe/savetoink/backend/lib/service/servicetypes"
	"github.com/spf13/cobra"
)

const defaultTimeoutSeconds = 30

var (
	outputPath string
	timeout    time.Duration
	verbose    bool

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

func runConvert(_ *cobra.Command, args []string) error {
	url := args[0]

	cfg, err := config.Load(consts.ModeCLI, nil)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if sendEmail {
		if cfg.EmailProvider != consts.EmailBackendMailjet {
			return fmt.Errorf("missing or unsupported email provider: '%s'", cfg.EmailProvider)
		}

		var missing []string
		cfg.ValidateEmailProviderConfigCli(&missing)

		if len(missing) > 0 {
			return fmt.Errorf("missing email provider config: %s", strings.Join(missing, ", "))
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Printf("Fetching article from: %s\n", url)

	svc := service.NewFromConfig(cfg)

	start := time.Now()
	result, err := svc.Process(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to process article: %w", err)
	}
	fmt.Printf("Processed in %v\n", time.Since(start))

	printVerboseOutput(result)

	var resp *email.SendEmailResponse
	if sendEmail {
		resp, err = svc.SendArticle(ctx, result.Article(), "", destEmail)
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
	} else if outputPath != "" {
		if writeErr := svc.WriteToFile(result, outputPath); writeErr != nil {
			return fmt.Errorf("failed to write EPUB: %w", writeErr)
		}
	}

	printResult(resp)

	return nil
}

func printVerboseOutput(result *servicetypes.ProcessResult) {
	if verbose {
		fmt.Println("\n--- Extracted Content (HTML) ---")
		fmt.Println(result.Article().Content)
		fmt.Println("--- End of Extracted Content ---")
		fmt.Println()
	}
}

func printResult(resp *email.SendEmailResponse) {
	if !sendEmail {
		if outputPath == "" {
			outputPath = "article.epub"
		}
		absPath, _ := filepath.Abs(outputPath)
		fmt.Printf("\n✓ EPUB saved to: %s\n", absPath)
	} else if resp != nil {
		fmt.Printf("\n✓ Article sent to Kindle (email ID: %s)\n", resp.MessageID)
	}
}

func main() {
	convertCmd.Flags().StringVarP(&outputPath, "output", "o", "article.epub", "Output file path")
	convertCmd.Flags().DurationVarP(&timeout, "timeout", "t",
		defaultTimeoutSeconds*time.Second, "Timeout for HTTP requests")
	convertCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show extracted HTML content")

	convertCmd.Flags().BoolVar(&sendEmail, "send", false, "Send EPUB to Kindle via email instead of saving locally")
	convertCmd.Flags().StringVar(&destEmail, "dest-email", "", "Destination Kindle email address")

	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
