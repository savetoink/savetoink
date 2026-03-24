package repository

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

const (
	attributeNameAccount         = "account"
	attributeNameID              = "id"
	attributeNameFavorite        = "favorite"
	attributeNameAccountFavorite = "accountFavorite"
)

// DynamoDB implements Repository interface using AWS DynamoDB.
type DynamoDB struct {
	client               *dynamodb.Client
	articleTableName     string
	articleTagsTableName string
	profileTableName     string
	sendsTableName       string
}

// NewDynamoDB creates a new DynamoDB repository instance.
func NewDynamoDB(
	awsConfig *aws.Config,
	articlesTableName, profileTableName, sendsTableName, articleTagsTableName string,
) *DynamoDB {
	if articlesTableName == "" {
		panic("articles table name is required")
	}
	if profileTableName == "" {
		panic("user profile table name is required")
	}
	if sendsTableName == "" {
		panic("sends table name is required")
	}

	cfg, _ := config.LoadDefaultConfig(context.TODO())
	if awsConfig != nil && awsConfig.Region == "" {
		cfg.Region = awsConfig.Region
	}
	return &DynamoDB{
		client:               dynamodb.NewFromConfig(cfg),
		articleTableName:     articlesTableName,
		articleTagsTableName: articleTagsTableName,
		profileTableName:     profileTableName,
		sendsTableName:       sendsTableName,
	}
}
