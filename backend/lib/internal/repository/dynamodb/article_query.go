package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	internaltypes "github.com/shaftoe/savetoink/backend/lib/internal/types"
	"github.com/shaftoe/savetoink/backend/lib/model"
)

const (
	attrNameAccount         = "#account"
	attrNameAccountFavorite = "#af = :accountFavorite"
)

func (d *DynamoDB) getProjectionAttributeNames() map[string]string {
	return map[string]string{
		"#a":    "account",
		"#i":    "id",
		"#u":    "url",
		"#c":    "createdAt",
		"#t":    "title",
		"#au":   "author",
		"#sn":   "siteName",
		"#sd":   "sourceDomain",
		"#e":    "excerpt",
		"#iurl": "imageUrl",
		"#l":    "language",
		"#err":  "error",
		"#wc":   "wordCount",
		"#rt":   "readingTimeMinutes",
		"#p":    "publishedAt",
		"#tg":   "tags",
		"#ds":   "deliveryStatus",
		"#af":   "accountFavorite",
	}
}

func (d *DynamoDB) getProjectionExpression(attrNames map[string]string) string {
	keys := make([]string, 0, len(attrNames))
	for key := range attrNames {
		if key != "#account" {
			keys = append(keys, key)
		}
	}

	return strings.Join(keys, ", ")
}

func (d *DynamoDB) totalCountByAccount(ctx context.Context, account string, favoriteFilter *bool) (int, error) {
	attrNames := map[string]string{}
	attrValues := map[string]types.AttributeValue{}

	var indexName string
	var keyConditionExpression string

	if favoriteFilter != nil && *favoriteFilter {
		// Use sparse GSI for favorites
		attrNames["#af"] = attributeNameAccountFavorite
		attrValues[":accountFavorite"] = &types.AttributeValueMemberS{Value: account + "#favorite"}
		indexName = consts.DynamoDBAccountFavoriteIndex
		keyConditionExpression = attrNameAccountFavorite
	} else {
		// Use regular GSI for all articles
		attrNames[attrNameAccount] = attributeNameAccount
		attrValues[":account"] = &types.AttributeValueMemberS{Value: account}
		indexName = consts.DynamoDBGSIName
		keyConditionExpression = attrNameAccount + " = :account"
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(d.articleTableName),
		IndexName:                 aws.String(indexName),
		KeyConditionExpression:    aws.String(keyConditionExpression),
		ExpressionAttributeNames:  attrNames,
		ExpressionAttributeValues: attrValues,
		Select:                    types.SelectCount,
	}

	resp, err := d.client.Query(ctx, queryInput)
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
	filter *internaltypes.ArticleFilter,
) (articles []*model.Article, total int, err error) {
	if page < consts.MinPage || pageSize < consts.MinPageSize || pageSize > consts.MaxPageSize {
		pageSize = consts.DefaultPageSize
	}

	var favoriteFilter *bool
	if filter != nil {
		favoriteFilter = filter.Favorite
	}

	total, err = d.totalCountByAccount(ctx, account, favoriteFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get count: %w", err)
	}

	if total == 0 {
		return []*model.Article{}, 0, nil
	}

	offset := (page - 1) * pageSize
	if offset >= total {
		return []*model.Article{}, total, nil
	}

	var exclusiveStartKey map[string]types.AttributeValue
	var resp *dynamodb.QueryOutput

	for i := 0; i < offset; i += pageSize {
		skipSize := min(pageSize, offset-i)
		resp, err = d.queryArticlesByAccount(ctx, account, skipSize, exclusiveStartKey, favoriteFilter)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to query articles: %w", err)
		}
		exclusiveStartKey = resp.LastEvaluatedKey
		if exclusiveStartKey == nil {
			break
		}
	}

	resp, err = d.queryArticlesByAccount(ctx, account, pageSize, exclusiveStartKey, favoriteFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query articles: %w", err)
	}

	articles, err = d.unmarshalArticles(resp.Items)
	if err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

func (d *DynamoDB) queryArticlesByAccount(
	ctx context.Context,
	account string,
	pageSize int,
	exclusiveStartKey map[string]types.AttributeValue,
	favoriteFilter *bool,
) (*dynamodb.QueryOutput, error) {
	attrNames := d.getProjectionAttributeNames()
	attrValues := map[string]types.AttributeValue{}

	var indexName string
	var keyConditionExpression string

	if favoriteFilter != nil && *favoriteFilter {
		// Use sparse GSI for favorites
		attrValues[":accountFavorite"] = &types.AttributeValueMemberS{Value: account + "#favorite"}
		attrNames["#af"] = attributeNameAccountFavorite
		indexName = consts.DynamoDBAccountFavoriteIndex
		keyConditionExpression = attrNameAccountFavorite
	} else {
		// Use regular GSI for all articles
		attrValues[":account"] = &types.AttributeValueMemberS{Value: account}
		attrNames[attrNameAccount] = attributeNameAccount
		indexName = consts.DynamoDBGSIName
		keyConditionExpression = attrNameAccount + " = :account"
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(d.articleTableName),
		IndexName:                 aws.String(indexName),
		KeyConditionExpression:    aws.String(keyConditionExpression),
		ProjectionExpression:      aws.String(d.getProjectionExpression(attrNames)),
		ExpressionAttributeNames:  attrNames,
		ExpressionAttributeValues: attrValues,
		ScanIndexForward:          aws.Bool(false),
		Limit:                     aws.Int32(int32(pageSize)), //nolint:gosec // pageSize is validated to be <= 20
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
		if err := unmarshalItem(item, &article, "article"); err != nil {
			return nil, err
		}

		article.Favorite = isFavorite(item)
		articles = append(articles, &article)
	}
	return articles, nil
}
