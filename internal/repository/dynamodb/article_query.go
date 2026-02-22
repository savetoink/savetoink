package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shaftoe/savetoink/internal/consts"
	"github.com/shaftoe/savetoink/internal/model"
)

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
			"#a, #i, #u, #c, #t, #au, #sn, #sd, #e, #iurl, #ct, #l, #err, #wc, #rt, #p, #dst, #df, #dt, #deu, #db",
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
