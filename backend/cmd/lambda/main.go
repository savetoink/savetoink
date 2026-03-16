// Package main implements the Lambda entry point for the API server.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/akrylysov/algnhsa"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/server"
)

func main() {
	cfg, err := config.Load(consts.ModeServer, func(ctx context.Context) (aws.Config, error) {
		return awsconfig.LoadDefaultConfig(ctx)
	})
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	router := server.NewRouter(cfg)

	algnhsa.ListenAndServe(router, nil)
}
