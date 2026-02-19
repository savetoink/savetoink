// Package main implements the CLI application for converting web articles to EPUB.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/shaftoe/savetoink/internal/config"
	"github.com/shaftoe/savetoink/internal/consts"
	"github.com/shaftoe/savetoink/internal/email"
	"github.com/shaftoe/savetoink/internal/service"
	"github.com/spf13/cobra"
)

const defaultTimeoutSeconds = 30

var (
	outputPath string
	timeout    time.Duration
	verbose    bool

	sendEmail    bool
	emailSubject string
	destEmail    string
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

	cfg, err := config.Load(consts.ModeCLI)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if sendEmail {
		if sendErr := cfg.EnableEmailSending(); sendErr != nil {
			return fmt.Errorf("failed to enable email sending: %w", sendErr)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Printf("Fetching article from: %s\n", url)

	svc := service.New(cfg)

	start := time.Now()
	result, err := svc.Process(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to process article: %w", err)
	}
	fmt.Printf("Processed in %v\n", time.Since(start))

	printVerboseOutput(result)

	var resp *email.SendEmailResponse
	if sendEmail {
		resp, err = svc.Send(ctx, result, emailSubject, destEmail)
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

func printVerboseOutput(result *service.ProcessResult) {
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
		fmt.Printf("\n✓ Article sent to Kindle (email ID: %s)\n", resp.EmailUUID)
	}
}

func main() {
	convertCmd.Flags().StringVarP(&outputPath, "output", "o", "article.epub", "Output file path")
	convertCmd.Flags().DurationVarP(&timeout, "timeout", "t",
		defaultTimeoutSeconds*time.Second, "Timeout for HTTP requests")
	convertCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show extracted HTML content")

	convertCmd.Flags().BoolVar(&sendEmail, "send", false, "Send EPUB to Kindle via email instead of saving locally")
	convertCmd.Flags().StringVar(&destEmail, "dest-email", "", "Destination Kindle email address")
	convertCmd.Flags().StringVar(&emailSubject, "email-subject", "", "Email subject (defaults to article title)")

	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
