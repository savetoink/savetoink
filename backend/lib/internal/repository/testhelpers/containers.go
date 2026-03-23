// Package testhelpers provides utilities for integration testing.
package testhelpers

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	transport "github.com/aws/smithy-go/endpoints"
	tcdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
)

// CreateDynamoDBContainer starts a local DynamoDB container for testing.
func CreateDynamoDBContainer(ctx context.Context) (*tcdynamodb.DynamoDBContainer, error) {
	ctr, err := tcdynamodb.Run(ctx, "amazon/dynamodb-local:2.2.1",
		tcdynamodb.WithDisableTelemetry(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start dynamodb container: %w", err)
	}

	return ctr, nil
}

// CreateDynamoDBClient creates a DynamoDB client configured to connect to the local container.
func CreateDynamoDBClient(ctx context.Context, container *tcdynamodb.DynamoDBContainer) (*dynamodb.Client, error) {
	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoint: %w", err)
	}

	cfg := aws.Config{
		Region:      "us-east-1",
		Credentials: credentials.NewStaticCredentialsProvider("DUMMYIDEXAMPLE", "DUMMYEXAMPLEKEY", ""),
	}

	return dynamodb.NewFromConfig(cfg, dynamodb.WithEndpointResolverV2(&dynamoDBResolver{HostPort: endpoint})), nil
}

type dynamoDBResolver struct {
	HostPort string
}

func (r *dynamoDBResolver) ResolveEndpoint(_ context.Context, params dynamodb.EndpointParameters) ( //nolint: gocritic

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
