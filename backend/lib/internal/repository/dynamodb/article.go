// Package repository provides DynamoDB implementations for data persistence.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shaftoe/savetoink/backend/lib/model"
)

// Store saves an article to DynamoDB.
func (d *DynamoDB) Store(ctx context.Context, article *model.Article) error {
	if article.Account == "" {
		return errors.New("account field is required")
	}

	item, err := attributevalue.MarshalMap(article)
	if err != nil {
		return fmt.Errorf("failed to marshal article: %w", err)
	}

	// If the article is a favorite, set the accountFavorite attribute for the sparse index
	if article.Favorite {
		item[attributeNameAccountFavorite] = &types.AttributeValueMemberS{Value: article.Account + "#favorite"}
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

	article.Favorite = isFavorite(resp.Item)

	return &article, nil
}

func isFavorite(item map[string]types.AttributeValue) bool {
	// set favorite based on presence of accountFavorite attribute
	return item[attributeNameAccountFavorite] != nil
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
	attrNames := map[string]string{
		"#af": attributeNameAccountFavorite,
	}

	var updateExpr string
	var attrValues map[string]types.AttributeValue

	if favorite {
		// When toggling on: SET #af = :accountFavorite
		attrValues = map[string]types.AttributeValue{
			":accountFavorite": &types.AttributeValueMemberS{Value: account + "#favorite"},
		}
		updateExpr = "SET #af = :accountFavorite"
	} else {
		// When toggling off: REMOVE #af
		updateExpr = "REMOVE #af"
	}

	_, err := d.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(d.articleTableName),
		Key: map[string]types.AttributeValue{
			attributeNameAccount: &types.AttributeValueMemberS{Value: account},
			attributeNameID:      &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeNames:  attrNames,
		ExpressionAttributeValues: attrValues,
	})
	if err != nil {
		return fmt.Errorf("failed to update favorite: %w", err)
	}

	return nil
}
