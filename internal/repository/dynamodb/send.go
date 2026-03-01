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
	"github.com/shaftoe/savetoink/internal/model"
)

// CreateSend implements SendsRepository.CreateSend.
func (d *DynamoDB) CreateSend(ctx context.Context, send *model.Send) error {
	if d.sendsTableName == "" {
		return errors.New("sends table not configured")
	}

	if send.PK == "" {
		return errors.New("pk field is required")
	}

	if send.SK == "" {
		return errors.New("sk field is required")
	}

	item, err := attributevalue.MarshalMap(send)
	if err != nil {
		return fmt.Errorf("failed to marshal send: %w", err)
	}

	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.sendsTableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to store send: %w", err)
	}

	return nil
}

// GetSendsByArticleID implements SendsRepository.GetSendsByArticleID.
func (d *DynamoDB) GetSendsByArticleID(ctx context.Context, articleID string) ([]*model.Send, error) {
	if d.sendsTableName == "" {
		return nil, errors.New("sends table not configured")
	}

	resp, err := d.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(d.sendsTableName),
		IndexName:              aws.String("ArticleIdIndex"),
		KeyConditionExpression: aws.String("articleId = :articleId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":articleId": &types.AttributeValueMemberS{Value: articleID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query sends by article id: %w", err)
	}

	if len(resp.Items) == 0 {
		return []*model.Send{}, nil
	}

	var sends []*model.Send
	for _, item := range resp.Items {
		var send model.Send
		if unmarshalErr := attributevalue.UnmarshalMap(item, &send); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal send: %w", unmarshalErr)
		}
		sends = append(sends, &send)
	}

	return sends, nil
}

// GetSendsByAccountDateRange implements SendsRepository.GetSendsByAccountDateRange.
func (d *DynamoDB) GetSendsByAccountDateRange(
	ctx context.Context,
	account string,
	startDate, endDate time.Time,
) ([]*model.Send, error) {
	if d.sendsTableName == "" {
		return nil, errors.New("sends table not configured")
	}

	resp, err := d.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(d.sendsTableName),
		IndexName:              aws.String("AccountSentAtIndex"),
		KeyConditionExpression: aws.String("account = :account AND sentAt BETWEEN :start AND :end"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":account": &types.AttributeValueMemberS{Value: account},
			":start":   &types.AttributeValueMemberS{Value: startDate.UTC().Format(time.RFC3339)},
			":end":     &types.AttributeValueMemberS{Value: endDate.UTC().Format(time.RFC3339)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query sends by account and date range: %w", err)
	}

	if len(resp.Items) == 0 {
		return []*model.Send{}, nil
	}

	var sends []*model.Send
	for _, item := range resp.Items {
		var send model.Send
		if unmarshalErr := attributevalue.UnmarshalMap(item, &send); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal send: %w", unmarshalErr)
		}
		sends = append(sends, &send)
	}

	return sends, nil
}
