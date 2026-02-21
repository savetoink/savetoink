// Package repository provides article tag index persistence implementation.
package repository

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shaftoe/savetoink/internal/consts"
)

const (
	tagIndexSKSeparator    = "#"
	tagIndexGSIName        = "GSI1"
	tagIndexMaxResultsSize = 100
)

// ArticleTags implements ArticleTagsRepository using AWS DynamoDB.
type ArticleTags struct {
	client *dynamodb.Client
	table  string
}

// NewArticleTags creates a new ArticleTags repository instance.
func NewArticleTags(awsConfig *aws.Config, tableName string) *ArticleTags {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	if awsConfig != nil && awsConfig.Region == "" {
		cfg.Region = awsConfig.Region
	}
	return &ArticleTags{
		client: dynamodb.NewFromConfig(cfg),
		table:  tableName,
	}
}

// AddTagIndex adds a tag index item for an article.
func (a *ArticleTags) AddTagIndex(ctx context.Context, account, articleID, tag, createdAt string) error {
	if account == "" || articleID == "" || tag == "" || createdAt == "" {
		return fmt.Errorf("account, articleID, tag, and createdAt are required")
	}

	normalizedTag := normalizeTag(tag)
	if normalizedTag == "" {
		return fmt.Errorf("tag cannot be empty or whitespace only")
	}

	pk := tagIndexPK(account, normalizedTag)
	sk := tagIndexSK(createdAt, articleID)

	item, err := attributevalue.MarshalMap(map[string]interface{}{
		"TagAccountKey":  pk,
		"ArticleSortKey": sk,
		"account":        account,
		"articleId":      articleID,
		"tag":            normalizedTag,
		"createdAt":      createdAt,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal tag index: %w", err)
	}

	_, err = a.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(a.table),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})
	if err != nil {
		return fmt.Errorf("failed to store tag index: %w", err)
	}

	return nil
}

// RemoveTagIndex removes a tag index item for an article.
func (a *ArticleTags) RemoveTagIndex(ctx context.Context, account, articleID, tag, createdAt string) error {
	if account == "" || articleID == "" || tag == "" || createdAt == "" {
		return fmt.Errorf("account, articleID, tag, and createdAt are required")
	}

	normalizedTag := normalizeTag(tag)
	if normalizedTag == "" {
		return fmt.Errorf("tag cannot be empty or whitespace only")
	}

	pk := tagIndexPK(account, normalizedTag)
	sk := tagIndexSK(createdAt, articleID)

	_, err := a.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(a.table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: sk},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete tag index: %w", err)
	}

	return nil
}

// GetArticleIDsByTag retrieves article IDs for a given tag with pagination.
func (a *ArticleTags) GetArticleIDsByTag(ctx context.Context, account, tag string, page, pageSize int) ([]string, map[string]types.AttributeValue, int, error) {
	if account == "" || tag == "" {
		return nil, nil, 0, fmt.Errorf("account and tag are required")
	}

	normalizedTag := normalizeTag(tag)
	if normalizedTag == "" {
		return nil, nil, 0, fmt.Errorf("tag cannot be empty or whitespace only")
	}

	if page < consts.MinPage || pageSize < consts.MinPageSize || pageSize > consts.MaxPageSize {
		pageSize = consts.DefaultPageSize
	}

	pk := tagIndexPK(account, normalizedTag)

	total, err := a.countByTag(ctx, pk)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get count: %w", err)
	}

	if total == 0 {
		return []string{}, nil, 0, nil
	}

	offset := (page - 1) * pageSize
	if offset >= total {
		return []string{}, nil, total, nil
	}

	var exclusiveStartKey map[string]types.AttributeValue
	var resp *dynamodb.QueryOutput

	for i := 0; i < offset; i += pageSize {
		skipSize := min(pageSize, offset-i)
		resp, err = a.queryByTag(ctx, pk, skipSize, exclusiveStartKey)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("failed to query tag index: %w", err)
		}
		exclusiveStartKey = resp.LastEvaluatedKey
		if exclusiveStartKey == nil {
			break
		}
	}

	resp, err = a.queryByTag(ctx, pk, pageSize, exclusiveStartKey)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to query tag index: %w", err)
	}

	articleIDs := make([]string, 0, len(resp.Items))
	for _, item := range resp.Items {
		var tagIndex struct {
			ArticleID string `dynamodbav:"articleId"`
		}
		if unmarshalErr := attributevalue.UnmarshalMap(item, &tagIndex); unmarshalErr != nil {
			continue
		}
		articleIDs = append(articleIDs, tagIndex.ArticleID)
	}

	return articleIDs, resp.LastEvaluatedKey, total, nil
}

// GetArticlesByTags retrieves article IDs that have all the specified tags (AND logic).
func (a *ArticleTags) GetArticlesByTags(ctx context.Context, account string, tags []string, page, pageSize int) ([]string, int, error) {
	if account == "" || len(tags) == 0 {
		return nil, 0, fmt.Errorf("account and tags are required")
	}

	normalizedTags := make([]string, 0, len(tags))
	for _, tag := range tags {
		nt := normalizeTag(tag)
		if nt != "" {
			normalizedTags = append(normalizedTags, nt)
		}
	}

	if len(normalizedTags) == 0 {
		return nil, 0, fmt.Errorf("tags cannot be empty or whitespace only")
	}

	sort.Strings(normalizedTags)

	articleIDsMap := make(map[string]bool)
	firstTag := true
	var allArticleIDs []string

	for _, tag := range normalizedTags {
		ids, _, _, err := a.GetArticleIDsByTag(ctx, account, tag, 1, tagIndexMaxResultsSize)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get article IDs for tag %s: %w", tag, err)
		}

		if firstTag {
			for _, id := range ids {
				articleIDsMap[id] = true
				allArticleIDs = append(allArticleIDs, id)
			}
			firstTag = false
		} else {
			newMap := make(map[string]bool)
			for _, id := range ids {
				if articleIDsMap[id] {
					newMap[id] = true
				}
			}
			articleIDsMap = newMap
		}

		if len(articleIDsMap) == 0 {
			return []string{}, 0, nil
		}
	}

	result := make([]string, 0, len(articleIDsMap))
	for id := range articleIDsMap {
		result = append(result, id)
	}

	sort.Strings(result)

	total := len(result)
	if total == 0 {
		return []string{}, 0, nil
	}

	offset := (page - 1) * pageSize
	if offset >= total {
		return []string{}, total, nil
	}

	end := min(offset+pageSize, total)
	return result[offset:end], total, nil
}

func (a *ArticleTags) countByTag(ctx context.Context, pk string) (int, error) {
	resp, err := a.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(a.table),
		IndexName:              aws.String(tagIndexGSIName),
		KeyConditionExpression: aws.String("GSI1PK = :gsi1pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi1pk": &types.AttributeValueMemberS{Value: pk},
		},
		Select: types.SelectCount,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to query count: %w", err)
	}
	return int(resp.Count), nil
}

func (a *ArticleTags) queryByTag(
	ctx context.Context,
	pk string,
	pageSize int,
	exclusiveStartKey map[string]types.AttributeValue,
) (*dynamodb.QueryOutput, error) {
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(a.table),
		IndexName:              aws.String(tagIndexGSIName),
		KeyConditionExpression: aws.String("GSI1PK = :gsi1pk"),
		ProjectionExpression:   aws.String("articleId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi1pk": &types.AttributeValueMemberS{Value: pk},
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int32(int32(pageSize)),
	}

	if exclusiveStartKey != nil {
		queryInput.ExclusiveStartKey = exclusiveStartKey
	}

	resp, err := a.client.Query(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to query dynamodb: %w", err)
	}
	return resp, nil
}

func tagIndexPK(account, tag string) string {
	return fmt.Sprintf("%s#%s", account, tag)
}

func tagIndexSK(createdAt, articleID string) string {
	return fmt.Sprintf("%s%s%s", createdAt, tagIndexSKSeparator, articleID)
}

func normalizeTag(tag string) string {
	tag = strings.TrimSpace(tag)
	tag = strings.ToLower(tag)
	if len(tag) > 50 {
		tag = tag[:50]
	}
	return tag
}
