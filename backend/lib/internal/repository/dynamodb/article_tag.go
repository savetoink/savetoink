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
	"github.com/shaftoe/savetoink/backend/lib/internal/validation"
	"github.com/shaftoe/savetoink/backend/lib/model"
)

const (
	attributeNameAccountTag         = "accountTag"
	attributeNameTag                = "tag"
	attributeNameArticleID          = "articleId"
	attributeNameCreatedAt          = "createdAt"
	attributeNameCreatedAtArticleID = "createdAtArticleId"

	//nolint:revive // GSI name matches CloudFormation template
	indexNameArticleIdCreatedAtIndex = "ArticleIdCreatedAtIndex"
	indexNameAccountTagIndex         = "AccountTagIndex"

	batchSize   = 25
	maxPageSize = 100
)

// getArticleCreatedAt retrieves the creation time for an article.
// Returns an error if the article doesn't exist.
func (d *DynamoDB) getArticleCreatedAt(
	ctx context.Context,
	accountID, articleID string,
) (time.Time, error) {
	article, err := d.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return time.Time{}, fmt.Errorf("article not found: %s", articleID)
		}
		return time.Time{}, fmt.Errorf("failed to get article: %w", err)
	}
	return article.CreatedAt, nil
}

// createPutWriteRequest creates a put write request for an article tag.
func createPutWriteRequest(
	accountID, articleID, tag string,
	createdAt time.Time,
) (types.WriteRequest, error) {
	item := model.ArticleTag{
		Account:            accountID,
		Tag:                tag,
		AccountTag:         buildAccountTagKey(accountID, tag),
		ArticleID:          articleID,
		CreatedAt:          createdAt,
		CreatedAtArticleID: buildCreatedAtArticleIDKey(createdAt, articleID),
	}

	marshaled, err := attributevalue.MarshalMap(item)
	if err != nil {
		return types.WriteRequest{}, fmt.Errorf("failed to marshal article tag: %w", err)
	}

	return types.WriteRequest{
		PutRequest: &types.PutRequest{
			Item: marshaled,
		},
	}, nil
}

// executeBatchWrite writes items to DynamoDB in batches.
func (d *DynamoDB) executeBatchWrite(
	ctx context.Context,
	writeReqs []types.WriteRequest,
) error {
	for i := 0; i < len(writeReqs); i += batchSize {
		_, err := d.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				d.articleTagsTableName: writeReqs[i:min(i+batchSize, len(writeReqs))],
			},
		})
		if err != nil {
			return fmt.Errorf("failed to batch write article tags: %w", err)
		}
	}
	return nil
}

// AddTagsToArticle adds tags to an article. If createdAt is nil, it fetches the article
// from the database to get the creation time. If createdAt is provided, it uses it directly,
// avoiding the extra database query.
func (d *DynamoDB) AddTagsToArticle(
	ctx context.Context,
	accountID, articleID string,
	tags []string,
	createdAt *time.Time,
) error {
	if len(tags) == 0 {
		return nil
	}

	// Deduplicate tags to prevent duplicate entries in the database
	tags = validation.DeduplicateStrings(tags)

	// Get the article creation time
	var articleCreatedAt time.Time
	if createdAt != nil {
		articleCreatedAt = *createdAt
	} else {
		var err error
		articleCreatedAt, err = d.getArticleCreatedAt(ctx, accountID, articleID)
		if err != nil {
			return err
		}
	}

	// Create write requests for batch operation
	writeReqs := make([]types.WriteRequest, 0, len(tags))
	for _, tag := range tags {
		req, err := createPutWriteRequest(accountID, articleID, tag, articleCreatedAt)
		if err != nil {
			return err
		}
		writeReqs = append(writeReqs, req)
	}

	// Execute batch write
	return d.executeBatchWrite(ctx, writeReqs)
}

// RemoveTagsFromArticle removes specific tags from an article.
func (d *DynamoDB) RemoveTagsFromArticle(ctx context.Context, accountID, articleID string, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	// Get the article's creation time for the range key
	article, err := d.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("article not found: %s", articleID)
		}
		return fmt.Errorf("failed to get article: %w", err)
	}

	// Create delete requests for batch operation
	writeReqs := make([]types.WriteRequest, 0, len(tags))
	createdAtKey := buildCreatedAtArticleIDKey(article.CreatedAt, articleID)
	for _, tag := range tags {
		accountTag := buildAccountTagKey(accountID, tag)

		deleteKey, marshalErr := attributevalue.MarshalMap(map[string]any{
			attributeNameAccountTag:         accountTag,
			attributeNameCreatedAtArticleID: createdAtKey,
		})
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal delete key: %w", marshalErr)
		}

		writeReqs = append(writeReqs, types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{
				Key: deleteKey,
			},
		})
	}

	// Execute batch write in chunks of batchSize (DynamoDB limit)
	for i := 0; i < len(writeReqs); i += batchSize {
		_, err = d.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				d.articleTagsTableName: writeReqs[i:min(i+batchSize, len(writeReqs))],
			},
		})
		if err != nil {
			return fmt.Errorf("failed to batch delete article tags: %w", err)
		}
	}

	return nil
}

// SetArticleTags replaces all tags for an article with the provided tags.
func (d *DynamoDB) SetArticleTags(ctx context.Context, accountID, articleID string, tags []string) error {
	// Get the article's creation time for the range key
	article, err := d.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("article not found: %s", articleID)
		}
		return fmt.Errorf("failed to get article: %w", err)
	}

	// First, delete all existing tags for this article
	err = d.DeleteTagsForArticle(ctx, accountID, articleID)
	if err != nil {
		return fmt.Errorf("failed to delete existing tags: %w", err)
	}

	// Then add the new tags
	return d.AddTagsToArticle(ctx, accountID, articleID, tags, &article.CreatedAt)
}

// GetArticleTags retrieves all tags for a specific article.
// GetArticleTags retrieves all tags for a specific article.
//
//nolint:gocritic // unused parameter required for interface compatibility
func (d *DynamoDB) GetArticleTags(ctx context.Context, _ string, articleID string) ([]string, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(d.articleTagsTableName),
		IndexName:              aws.String(indexNameArticleIdCreatedAtIndex),
		KeyConditionExpression: aws.String(attributeNameArticleID + " = :articleId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":articleId": &types.AttributeValueMemberS{Value: articleID},
		},
	}

	result, err := d.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query article tags: %w", err)
	}

	tags := make([]string, 0, len(result.Items))
	for _, item := range result.Items {
		var tag model.ArticleTag
		if unmarshalErr := attributevalue.UnmarshalMap(item, &tag); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal article tag: %w", unmarshalErr)
		}
		tags = append(tags, tag.Tag)
	}

	return tags, nil
}

// GetArticlesByTag retrieves article IDs for articles with a specific tag for a given account.
// Supports pagination with page (1-indexed) and pageSize parameters.
//
//nolint:funlen,gocritic // Function is moderately long; unnamed results follow repository pattern
func (d *DynamoDB) GetArticlesByTag(
	ctx context.Context,
	accountID, tag string,
	page, pageSize int,
) ([]string, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	input := &dynamodb.QueryInput{
		TableName:              aws.String(d.articleTagsTableName),
		KeyConditionExpression: aws.String(attributeNameAccountTag + " = :accountTag"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":accountTag": &types.AttributeValueMemberS{Value: buildAccountTagKey(accountID, tag)},
		},
		ScanIndexForward: aws.Bool(false), // Sort by createdAt descending
		Limit:            aws.Int32(int32(pageSize)),
	}

	// Handle pagination
	if page > 1 {
		// Get the last key from previous page
		// For simplicity, we'll fetch all items up to this page
		input.Limit = nil // Remove limit to get all items
		allItems, err := d.client.Query(ctx, input)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to query articles by tag: %w", err)
		}

		total := len(allItems.Items)
		start := (page - 1) * pageSize
		end := start + pageSize
		if start >= len(allItems.Items) {
			return []string{}, total, nil
		}
		if end > len(allItems.Items) {
			end = len(allItems.Items)
		}

		articles := make([]string, 0, end-start)
		for _, item := range allItems.Items[start:end] {
			var tag model.ArticleTag
			if unmarshalErr := attributevalue.UnmarshalMap(item, &tag); unmarshalErr != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal article tag: %w", unmarshalErr)
			}
			articles = append(articles, tag.ArticleID)
		}

		return articles, total, nil
	}

	result, err := d.client.Query(ctx, input)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query articles by tag: %w", err)
	}

	articles := make([]string, 0, len(result.Items))
	for _, item := range result.Items {
		var tag model.ArticleTag
		if unmarshalErr := attributevalue.UnmarshalMap(item, &tag); unmarshalErr != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal article tag: %w", unmarshalErr)
		}
		articles = append(articles, tag.ArticleID)
	}

	// Get total count by querying without limit
	total, err := d.getTotalCountForTag(ctx, accountID, tag)
	if err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// getTotalCountForTag gets the total count of articles with a specific tag.
func (d *DynamoDB) getTotalCountForTag(ctx context.Context, accountID, tag string) (int, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(d.articleTagsTableName),
		KeyConditionExpression: aws.String(attributeNameAccountTag + " = :accountTag"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":accountTag": &types.AttributeValueMemberS{Value: buildAccountTagKey(accountID, tag)},
		},
		Select: types.SelectCount,
	}

	result, err := d.client.Query(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to get count for tag: %w", err)
	}

	return int(result.Count), nil
}

// GetAllTagsForAccount retrieves all unique tags for an account.
func (d *DynamoDB) GetAllTagsForAccount(ctx context.Context, accountID string) ([]string, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(d.articleTagsTableName),
		IndexName:              aws.String(indexNameAccountTagIndex),
		KeyConditionExpression: aws.String(attributeNameAccount + " = :account"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":account": &types.AttributeValueMemberS{Value: accountID},
		},
	}

	result, err := d.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query all tags for account: %w", err)
	}

	// Use a map to deduplicate tags (same tag can appear on multiple articles)
	tagMap := make(map[string]bool)
	for _, item := range result.Items {
		var articleTag model.ArticleTag
		if unmarshalErr := attributevalue.UnmarshalMap(item, &articleTag); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal article tag: %w", unmarshalErr)
		}
		tagMap[articleTag.Tag] = true
	}

	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}

	return tags, nil
}

// DeleteTagsForArticle deletes all tags for a specific article.
func (d *DynamoDB) DeleteTagsForArticle(ctx context.Context, accountID, articleID string) error {
	// Get all tags for this article
	tags, err := d.GetArticleTags(ctx, accountID, articleID)
	if err != nil {
		return fmt.Errorf("failed to get article tags: %w", err)
	}

	if len(tags) == 0 {
		return nil
	}

	// Get the article's creation time for the range key
	article, err := d.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("article not found: %s", articleID)
		}
		return fmt.Errorf("failed to get article: %w", err)
	}

	// Create delete requests for batch operation
	writeReqs := make([]types.WriteRequest, 0, len(tags))
	createdAtKey := buildCreatedAtArticleIDKey(article.CreatedAt, articleID)
	for _, tag := range tags {
		accountTag := buildAccountTagKey(accountID, tag)

		deleteKey, marshalErr := attributevalue.MarshalMap(map[string]any{
			attributeNameAccountTag:         accountTag,
			attributeNameCreatedAtArticleID: createdAtKey,
		})
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal delete key: %w", marshalErr)
		}

		writeReqs = append(writeReqs, types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{
				Key: deleteKey,
			},
		})
	}

	// Execute batch write in chunks of batchSize (DynamoDB limit)
	for i := 0; i < len(writeReqs); i += batchSize {
		_, err = d.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				d.articleTagsTableName: writeReqs[i:min(i+batchSize, len(writeReqs))],
			},
		})
		if err != nil {
			return fmt.Errorf("failed to batch delete article tags: %w", err)
		}
	}

	return nil
}

// Helper functions for building DynamoDB keys

func buildAccountTagKey(accountID, tag string) string {
	return fmt.Sprintf("%s:%s", accountID, tag)
}

func buildCreatedAtArticleIDKey(createdAt time.Time, articleID string) string {
	return fmt.Sprintf("%d:%s", createdAt.Unix(), articleID)
}
