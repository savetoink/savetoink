// Package testhelpers provides utilities for integration testing.
package testhelpers

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	transport "github.com/aws/smithy-go/endpoints"
	tclocalstack "github.com/testcontainers/testcontainers-go/modules/localstack"
)

type lambdaResolver struct {
	HostPort string
}

func (r *lambdaResolver) ResolveEndpoint(_ context.Context, params lambda.EndpointParameters) (
	transport.Endpoint,
	error,
) {
	_ = params
	return transport.Endpoint{
		URI: url.URL{
			Scheme: "http",
			Host:   r.HostPort,
		},
	}, nil
}

// CreateLocalStackContainer starts a localstack container.
func CreateLocalStackContainer(ctx context.Context) (*tclocalstack.LocalStackContainer, error) {
	ctr, err := tclocalstack.Run(ctx, "localstack/localstack:latest")
	if err != nil {
		return nil, fmt.Errorf("failed to start localstack container: %w", err)
	}

	return ctr, nil
}

// CreateLambdaClient creates a lambda client connected to the localstack container.
func CreateLambdaClient(ctx context.Context, container *tclocalstack.LocalStackContainer) (*lambda.Client, error) {
	endpoint, err := container.Endpoint(ctx, "4566")
	if err != nil {
		return nil, fmt.Errorf("failed to get localstack endpoint: %w", err)
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("DUMMYIDEXAMPLE", "DUMMYEXAMPLEKEY", "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	return lambda.NewFromConfig(cfg, lambda.WithEndpointResolverV2(&lambdaResolver{HostPort: endpoint})), nil
}
