// Package testhelpers provides utilities for integration testing.
package testhelpers

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	accountCreatedAtIndexProjection = []string{
		"author",
		"deliveryStatus",
		"error",
		"excerpt",
		"favorite",
		"id",
		"imageUrl",
		"language",
		"publishedAt",
		"readingTimeMinutes",
		"siteName",
		"sourceDomain",
		"tags",
		"title",
		"url",
		"wordCount",
	}

	accountFavoriteIndexProjection = append([]string{"account"}, accountCreatedAtIndexProjection...)
)

const (
	errFailedCreateTable = "failed to create table: %w"
)

// SetupAllTables creates all required DynamoDB tables for testing.
func SetupAllTables(ctx context.Context, client *dynamodb.Client) error {
	if err := createArticleTable(ctx, client); err != nil {
		return fmt.Errorf(errFailedCreateTable, err)
	}
	if err := createArticleTagsTable(ctx, client); err != nil {
		return fmt.Errorf("failed to create article tags table: %w", err)
	}
	if err := createUserProfileTable(ctx, client); err != nil {
		return fmt.Errorf("failed to create user profile table: %w", err)
	}
	if err := createSendsTable(ctx, client); err != nil {
		return fmt.Errorf("failed to create sends table: %w", err)
	}
	return nil
}

func createArticleTable(ctx context.Context, client *dynamodb.Client) error {
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("test-articles"),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("account"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("createdAt"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("accountFavorite"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("account"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeRange},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("AccountCreatedAtIndex"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("account"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("createdAt"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{
					ProjectionType:   types.ProjectionTypeInclude,
					NonKeyAttributes: accountCreatedAtIndexProjection,
				},
			},
			{
				IndexName: aws.String("AccountFavoriteIndex"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("accountFavorite"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("createdAt"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{
					ProjectionType:   types.ProjectionTypeInclude,
					NonKeyAttributes: accountFavoriteIndexProjection,
				},
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		return fmt.Errorf(errFailedCreateTable, err)
	}

	return waitForTableActive(ctx, client, "test-articles")
}

func createArticleTagsTable(ctx context.Context, client *dynamodb.Client) error {
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("test-article-tags"),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("accountTag"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("createdAtArticleId"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("accountTag"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("createdAtArticleId"), KeyType: types.KeyTypeRange},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		return fmt.Errorf(errFailedCreateTable, err)
	}

	return waitForTableActive(ctx, client, "test-article-tags")
}

func createUserProfileTable(ctx context.Context, client *dynamodb.Client) error {
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("test-user-profiles"),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("account"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("deviceEmail"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("account"), KeyType: types.KeyTypeHash},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("DeviceEmailIndex"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("deviceEmail"), KeyType: types.KeyTypeHash},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		return fmt.Errorf(errFailedCreateTable, err)
	}

	return waitForTableActive(ctx, client, "test-user-profiles")
}

func createSendsTable(ctx context.Context, client *dynamodb.Client) error {
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("test-sends"),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("pk"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("sk"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("account"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("articleId"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("sentAt"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("pk"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("sk"), KeyType: types.KeyTypeRange},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("ArticleIdIndex"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("articleId"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("sentAt"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
			},
			{
				IndexName: aws.String("AccountSentAtIndex"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("account"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("sentAt"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		return fmt.Errorf(errFailedCreateTable, err)
	}

	return waitForTableActive(ctx, client, "test-sends")
}

func waitForTableActive(ctx context.Context, client *dynamodb.Client, tableName string) error {
	const maxAttempts = 20
	const waitTime = 500 * time.Millisecond

	for range maxAttempts {
		resp, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err != nil {
			return fmt.Errorf("failed to describe table %s: %w", tableName, err)
		}

		if resp.Table.TableStatus == types.TableStatusActive {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled: %w", ctx.Err())
		case <-time.After(waitTime):
		}
	}

	return fmt.Errorf("table %s did not become active in time", tableName)
}
