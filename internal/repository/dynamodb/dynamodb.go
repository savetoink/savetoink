package repository

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

const (
	attributeNameAccount  = "account"
	attributeNameID       = "id"
	attributeNameFavorite = "favorite"
)

// DynamoDB implements Repository interface using AWS DynamoDB.
type DynamoDB struct {
	client           *dynamodb.Client
	articleTableName string
	profileTableName string
}

// NewDynamoDB creates a new DynamoDB repository instance.
func NewDynamoDB(awsConfig *aws.Config, articlesTableName, profileTableName string) *DynamoDB {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	if awsConfig != nil && awsConfig.Region == "" {
		cfg.Region = awsConfig.Region
	}
	return &DynamoDB{
		client:           dynamodb.NewFromConfig(cfg),
		articleTableName: articlesTableName,
		profileTableName: profileTableName,
	}
}
