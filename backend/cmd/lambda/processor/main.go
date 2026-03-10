// Lambda processor handles async article processing.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	lambdahandler "github.com/shaftoe/savetoink/backend/lib/processor/lambda"
	"github.com/shaftoe/savetoink/backend/lib/service"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
)

func main() {
	cfg, err := config.Load(consts.ModeServer, func(ctx context.Context) (aws.Config, error) {
		return awsconfig.LoadDefaultConfig(ctx)
	})
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logging.SetupLogging(cfg)
	svc := service.NewFromConfig(cfg)
	lambda.Start(func(ctx context.Context, event *content.ProcessArticleEvent) error {
		return lambdahandler.HandleEvent(ctx, event, svc)
	})
}
