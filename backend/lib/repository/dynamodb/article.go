// Package repository provides DynamoDB implementations for data persistence.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/model"
)

// Store saves an article to DynamoDB.
func (d *DynamoDB) Store(ctx context.Context, article *model.Article) error {
	now := time.Now().UTC()

	if article.Account == "" {
		return errors.New("account field is required")
	}

	if article.CreatedAt.IsZero() {
		article.CreatedAt = now
	}

	item, err := attributevalue.MarshalMap(article)
	if err != nil {
		return fmt.Errorf("failed to marshal article: %w", err)
	}

	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.articleTableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to store article: %w", err)
	}

	return nil
}

// GetByAccountAndID implements Repository.GetByAccountAndID.
func (d *DynamoDB) GetByAccountAndID(ctx context.Context, account, id string) (*model.Article, error) {
	resp, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.articleTableName),
		Key: map[string]types.AttributeValue{
			attributeNameAccount: &types.AttributeValueMemberS{Value: account},
			attributeNameID:      &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	if resp.Item == nil {
		return nil, ErrNotFound
	}

	var article model.Article
	if unmarshalErr := unmarshalItem(resp.Item, &article, "article"); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &article, nil
}

// DeleteByAccountAndID implements Repository.DeleteByAccountAndID.
func (d *DynamoDB) DeleteByAccountAndID(ctx context.Context, account, id string) error {
	_, err := d.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(d.articleTableName),
		Key: map[string]types.AttributeValue{
			attributeNameAccount: &types.AttributeValueMemberS{Value: account},
			attributeNameID:      &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete article: %w", err)
	}

	return nil
}

// UpdateFavorite updates the favorite status of an article.
func (d *DynamoDB) UpdateFavorite(ctx context.Context, account, id string, favorite bool) error {
	_, err := d.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(d.articleTableName),
		Key: map[string]types.AttributeValue{
			attributeNameAccount: &types.AttributeValueMemberS{Value: account},
			attributeNameID:      &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression: aws.String("SET #f = :favorite"),
		ExpressionAttributeNames: map[string]string{
			"#f": "favorite",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":favorite": &types.AttributeValueMemberBOOL{Value: favorite},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update favorite: %w", err)
	}

	return nil
}

// DeleteByAccount implements Repository.DeleteByAccount.
func (d *DynamoDB) DeleteByAccount(ctx context.Context, account string) (int, error) {
	articles, _, _, err := d.GetMetadataByAccount(ctx, account, 1, consts.MaxPageSize, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get articles for deletion: %w", err)
	}

	if len(articles) == 0 {
		return 0, nil
	}

	for i := 0; i < len(articles); i += consts.DynamoDBBatchSize {
		end := min(i+consts.DynamoDBBatchSize, len(articles))

		var writeReqs []types.WriteRequest
		for _, article := range articles[i:end] {
			writeReqs = append(writeReqs, types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: map[string]types.AttributeValue{
						attributeNameAccount: &types.AttributeValueMemberS{Value: article.Account},
						attributeNameID:      &types.AttributeValueMemberS{Value: article.ID},
					},
				},
			})
		}

		_, err = d.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				d.articleTableName: writeReqs,
			},
		})
		if err != nil {
			return i, fmt.Errorf("failed to delete batch of articles: %w", err)
		}
	}

	return len(articles), nil
}
