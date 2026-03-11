// Package main implements the CLI application for converting web articles to EPUB.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/logging"
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

	cfg *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "savetoink",
	Short: "Convert web articles to EPUB format",
	Long:  `A CLI tool to fetch web articles and convert them to EPUB format for Kindle devices.`,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		var err error

		cfg, err = config.Load(consts.ModeCLI, nil)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		logging.SetupLogging(cfg)
		return nil
	},
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

var cleanCmd = &cobra.Command{
	Use:   "clean [url]",
	Short: "Fetch and clean a URL, print cleaned content to stdout",
	Args:  cobra.ExactArgs(1),
	RunE:  runClean,
}

func main() {
	convertCmd.Flags().StringVarP(&outputPath, "output", "o", "article.epub", "Output file path")
	convertCmd.Flags().DurationVarP(&timeout, "timeout", "t",
		defaultTimeoutSeconds*time.Second, "Timeout for HTTP requests")

	convertCmd.Flags().BoolVar(&sendEmail, "send", false, "Send EPUB to Kindle via email instead of saving locally")
	convertCmd.Flags().StringVar(&destEmail, "dest-email", "", "Destination Kindle email address")

	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
