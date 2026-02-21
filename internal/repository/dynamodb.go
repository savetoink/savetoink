// Package repository provides article persistence implementations.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shaftoe/savetoink/internal/consts"
	"github.com/shaftoe/savetoink/internal/model"
)

const (
	attributeNameAccount = "account"
	attributeNameID      = "id"
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
	if unmarshalErr := attributevalue.UnmarshalMap(resp.Item, &article); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal article: %w", unmarshalErr)
	}

	return &article, nil
}

func (d *DynamoDB) getProjectionAttributeNames() map[string]string {
	return map[string]string{
		"#account": attributeNameAccount,
		"#a":       "account",
		"#i":       "id",
		"#u":       "url",
		"#c":       "createdAt",
		"#t":       "title",
		"#au":      "author",
		"#sn":      "siteName",
		"#sd":      "sourceDomain",
		"#e":       "excerpt",
		"#iurl":    "imageUrl",
		"#ct":      "contentType",
		"#l":       "language",
		"#err":     "error",
		"#wc":      "wordCount",
		"#rt":      "readingTimeMinutes",
		"#p":       "publishedAt",
		"#tags":    "tags",
		"#dst":     "deliveryStatus",
		"#df":      "deliveredFrom",
		"#dt":      "deliveredTo",
		"#deu":     "deliveredEmailUUID",
		"#db":      "deliveredBy",
	}
}

func (d *DynamoDB) totalCountByAccount(ctx context.Context, account string) (int, error) {
	resp, err := d.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(d.articleTableName),
		IndexName:              aws.String(consts.DynamoDBGSIName),
		KeyConditionExpression: aws.String("#account = :account"),
		ExpressionAttributeNames: map[string]string{
			"#account": attributeNameAccount,
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":account": &types.AttributeValueMemberS{Value: account},
		},
		Select: types.SelectCount,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to query count: %w", err)
	}
	return int(resp.Count), nil
}

// GetMetadataByAccount implements Repository.GetMetadataByAccount.
// Returns articles with all metadata fields except content.
func (d *DynamoDB) GetMetadataByAccount(
	ctx context.Context,
	account string,
	page, pageSize int,
) (articles []*model.Article, lastEvaluatedKey map[string]types.AttributeValue, total int, err error) {
	if page < consts.MinPage || pageSize < consts.MinPageSize || pageSize > consts.MaxPageSize {
		pageSize = consts.DefaultPageSize
	}

	total, err = d.totalCountByAccount(ctx, account)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get count: %w", err)
	}

	if total == 0 {
		return []*model.Article{}, nil, 0, nil
	}

	offset := (page - 1) * pageSize
	if offset >= total {
		return []*model.Article{}, nil, total, nil
	}

	var exclusiveStartKey map[string]types.AttributeValue
	var resp *dynamodb.QueryOutput

	for i := 0; i < offset; i += pageSize {
		skipSize := min(pageSize, offset-i)
		resp, err = d.queryArticlesByAccount(ctx, account, skipSize, exclusiveStartKey)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("failed to query articles: %w", err)
		}
		exclusiveStartKey = resp.LastEvaluatedKey
		if exclusiveStartKey == nil {
			break
		}
	}

	resp, err = d.queryArticlesByAccount(ctx, account, pageSize, exclusiveStartKey)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to query articles: %w", err)
	}

	articles, err = d.unmarshalArticles(resp.Items)
	if err != nil {
		return nil, nil, 0, err
	}

	return articles, resp.LastEvaluatedKey, total, nil
}

func (d *DynamoDB) queryArticlesByAccount(
	ctx context.Context,
	account string,
	pageSize int,
	exclusiveStartKey map[string]types.AttributeValue,
) (*dynamodb.QueryOutput, error) {
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(d.articleTableName),
		IndexName:              aws.String(consts.DynamoDBGSIName),
		KeyConditionExpression: aws.String("#account = :account"),
		ProjectionExpression: aws.String(
			"#a, #i, #u, #c, #t, #au, #sn, #sd, #e, #iurl, #ct, #l, #err, #wc, #rt, #p, #tags, #dst, #df, #dt, #deu, #db",
		),
		ExpressionAttributeNames: d.getProjectionAttributeNames(),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":account": &types.AttributeValueMemberS{Value: account},
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int32(int32(pageSize)), //nolint:gosec // pageSize is validated to be <= 20
	}

	if exclusiveStartKey != nil {
		queryInput.ExclusiveStartKey = exclusiveStartKey
	}

	resp, err := d.client.Query(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to query dynamodb: %w", err)
	}
	return resp, nil
}

func (d *DynamoDB) unmarshalArticles(items []map[string]types.AttributeValue) ([]*model.Article, error) {
	articles := make([]*model.Article, 0, len(items))
	for _, item := range items {
		var article model.Article
		if unmarshalErr := attributevalue.UnmarshalMap(item, &article); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal article: %w", unmarshalErr)
		}
		articles = append(articles, &article)
	}
	return articles, nil
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

// DeleteByAccount implements Repository.DeleteByAccount.
func (d *DynamoDB) DeleteByAccount(ctx context.Context, account string) (int, error) {
	articles, _, _, err := d.GetMetadataByAccount(ctx, account, 1, consts.MaxPageSize)
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

// ErrNotFound is returned when an article is not found.
var ErrNotFound = errors.New("article not found")

// GetUserProfile implements UserProfileRepository.GetUserProfile.
func (d *DynamoDB) GetUserProfile(ctx context.Context, account string) (*model.UserProfile, error) {
	if d.profileTableName == "" {
		return nil, errors.New("user profile table not configured")
	}

	resp, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.profileTableName),
		Key: map[string]types.AttributeValue{
			attributeNameAccount: &types.AttributeValueMemberS{Value: account},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	if resp.Item == nil {
		return nil, nil
	}

	var profile model.UserProfile
	if unmarshalErr := attributevalue.UnmarshalMap(resp.Item, &profile); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal user profile: %w", unmarshalErr)
	}

	return &profile, nil
}

// PutUserProfile implements UserProfileRepository.PutUserProfile.
func (d *DynamoDB) PutUserProfile(ctx context.Context, profile *model.UserProfile) error {
	if d.profileTableName == "" {
		return errors.New("user profile table not configured")
	}

	if profile.Account == "" {
		return errors.New("account field is required")
	}

	item, err := attributevalue.MarshalMap(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal user profile: %w", err)
	}

	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.profileTableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to store user profile: %w", err)
	}

	return nil
}

// DeleteUserProfile implements UserProfileRepository.DeleteUserProfile.
func (d *DynamoDB) DeleteUserProfile(ctx context.Context, account string) error {
	if d.profileTableName == "" {
		return errors.New("user profile table not configured")
	}

	_, err := d.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(d.profileTableName),
		Key: map[string]types.AttributeValue{
			attributeNameAccount: &types.AttributeValueMemberS{Value: account},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete user profile: %w", err)
	}

	return nil
}
