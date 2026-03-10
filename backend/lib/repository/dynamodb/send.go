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

type dynamoDBSend struct {
	PK            string    `dynamodbav:"pk"`
	SK            string    `dynamodbav:"sk"`
	Account       string    `dynamodbav:"account"`
	ArticleID     string    `dynamodbav:"articleId"`
	SentAt        time.Time `dynamodbav:"sentAt"`
	Title         string    `dynamodbav:"title"`
	DestEmail     string    `dynamodbav:"destEmail"`
	Status        string    `dynamodbav:"status"`
	SenderEmail   string    `dynamodbav:"senderEmail"`
	MessageID     string    `dynamodbav:"messageID,omitempty"`
	Provider      string    `dynamodbav:"provider"`
	ErrorResponse string    `dynamodbav:"errorResponse,omitempty"`
}

func (d *dynamoDBSend) toDomain() *model.Send {
	return &model.Send{
		Account:       d.Account,
		ArticleID:     d.ArticleID,
		SentAt:        d.SentAt,
		Title:         d.Title,
		DestEmail:     d.DestEmail,
		Status:        d.Status,
		SenderEmail:   d.SenderEmail,
		MessageID:     d.MessageID,
		Provider:      d.Provider,
		ErrorResponse: d.ErrorResponse,
	}
}

// CreateSendRecord implements SendsRepository.CreateSendRecord.
func (d *DynamoDB) CreateSendRecord(ctx context.Context, send *model.Send) error {
	now := time.Now().UTC()
	dbSend := &dynamoDBSend{
		PK:          "USER#" + send.Account,
		SK:          "SEND#" + now.Format(time.RFC3339) + "#" + send.ArticleID,
		Account:     send.Account,
		ArticleID:   send.ArticleID,
		SentAt:      send.SentAt.UTC(),
		Title:       send.Title,
		DestEmail:   send.DestEmail,
		SenderEmail: send.SenderEmail,
		Provider:    send.Provider,
		Status:      "pending",
	}

	item, err := attributevalue.MarshalMap(dbSend)
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

// UpdateSendRecord implements SendsRepository.UpdateSendRecord.
func (d *DynamoDB) UpdateSendRecord(ctx context.Context, send *model.Send) error {
	dbSends, err := d.getDynamoDBSendsByArticleID(ctx, send.ArticleID)
	if err != nil {
		return fmt.Errorf("failed to get send: %w", err)
	}

	if len(dbSends) == 0 {
		return fmt.Errorf("send record not found for article %s", send.ArticleID)
	}

	dbSend := dbSends[len(dbSends)-1]
	if dbSend.Account != send.Account {
		return errors.New("send record account mismatch")
	}

	dbSend.Status = send.Status
	dbSend.MessageID = send.MessageID
	dbSend.ErrorResponse = send.ErrorResponse

	item, err := attributevalue.MarshalMap(dbSend)
	if err != nil {
		return fmt.Errorf("failed to marshal send: %w", err)
	}

	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.sendsTableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to update send: %w", err)
	}

	return nil
}

// GetSendsByArticleID implements SendsRepository.GetSendsByArticleID.
func (d *DynamoDB) GetSendsByArticleID(ctx context.Context, articleID string) ([]*model.Send, error) {
	dbSends, err := d.getDynamoDBSendsByArticleID(ctx, articleID)
	if err != nil {
		return nil, err
	}

	if len(dbSends) == 0 {
		return []*model.Send{}, nil
	}

	sends := make([]*model.Send, 0, len(dbSends))
	for _, dbSend := range dbSends {
		sends = append(sends, dbSend.toDomain())
	}

	return sends, nil
}

func (d *DynamoDB) getDynamoDBSendsByArticleID(ctx context.Context, articleID string) ([]*dynamoDBSend, error) {
	resp, err := d.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(d.sendsTableName),
		IndexName:              aws.String(consts.DynamoDBSendsArticleIDIndex),
		KeyConditionExpression: aws.String("articleId = :articleId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":articleId": &types.AttributeValueMemberS{Value: articleID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query sends by article id: %w", err)
	}

	if len(resp.Items) == 0 {
		return []*dynamoDBSend{}, nil
	}

	var sends []*dynamoDBSend
	for _, item := range resp.Items {
		var dbSend dynamoDBSend
		if unmarshalErr := attributevalue.UnmarshalMap(item, &dbSend); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal send: %w", unmarshalErr)
		}
		sends = append(sends, &dbSend)
	}

	return sends, nil
}

// GetSendsByAccountDateRange implements SendsRepository.GetSendsByAccountDateRange.
func (d *DynamoDB) GetSendsByAccountDateRange(
	ctx context.Context,
	account string,
	startDate, endDate time.Time,
) ([]*model.Send, error) {
	resp, err := d.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(d.sendsTableName),
		IndexName:              aws.String(consts.DynamoDBSendsAccountSentAtIndex),
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
		var dbSend dynamoDBSend
		if unmarshalErr := attributevalue.UnmarshalMap(item, &dbSend); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal send: %w", unmarshalErr)
		}
		sends = append(sends, dbSend.toDomain())
	}

	return sends, nil
}

// CountSendsByAccountDateRange counts the number of sends for a given account within a date range.
func (d *DynamoDB) CountSendsByAccountDateRange(
	ctx context.Context,
	account string,
	startDate, endDate time.Time,
) (int, error) {
	resp, err := d.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(d.sendsTableName),
		IndexName:              aws.String(consts.DynamoDBSendsAccountSentAtIndex),
		Select:                 types.SelectCount,
		KeyConditionExpression: aws.String("account = :account AND sentAt BETWEEN :start AND :end"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":account": &types.AttributeValueMemberS{Value: account},
			":start":   &types.AttributeValueMemberS{Value: startDate.UTC().Format(time.RFC3339)},
			":end":     &types.AttributeValueMemberS{Value: endDate.UTC().Format(time.RFC3339)},
		},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count sends: %w", err)
	}

	return int(resp.Count), nil
}
