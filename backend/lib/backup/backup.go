package backup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	maxScanLimit = 100
)

type DynamoDBScanner interface {
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

type S3Putter interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type BackupResult struct {
	TableName  string
	ItemsCount int
	Key        string
	Error      error
}

type BackupManager struct {
	dynamoDBClient DynamoDBScanner
	s3Client       S3Putter
	bucketName     string
	logger         *slog.Logger
}

func NewBackupManager(dynamoDBClient DynamoDBScanner, s3Client S3Putter, bucketName string, logger *slog.Logger) *BackupManager {
	return &BackupManager{
		dynamoDBClient: dynamoDBClient,
		s3Client:       s3Client,
		bucketName:     bucketName,
		logger:         logger,
	}
}

func (bm *BackupManager) BackupAllTables(ctx context.Context, tables []string) []BackupResult {
	results := make([]BackupResult, 0, len(tables))

	for _, tableName := range tables {
		result := bm.BackupTable(ctx, tableName)
		results = append(results, result)
	}

	return results
}

func (bm *BackupManager) BackupTable(ctx context.Context, tableName string) BackupResult {
	result := BackupResult{
		TableName: tableName,
	}

	if bm.bucketName == "" {
		result.Error = fmt.Errorf("backup bucket name is not configured")
		return result
	}

	items, err := bm.scanTable(ctx, tableName)
	if err != nil {
		result.Error = fmt.Errorf("failed to scan table %s: %w", tableName, err)
		return result
	}

	result.ItemsCount = len(items)

	key := fmt.Sprintf("backups/%s-%s.json", tableName, time.Now().UTC().Format("20060102-150405"))
	result.Key = key

	backupData := map[string]interface{}{
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"table_name": tableName,
		"item_count": len(items),
		"items":      items,
	}

	data, err := json.MarshalIndent(backupData, "", "  ")
	if err != nil {
		result.Error = fmt.Errorf("failed to marshal backup data for table %s: %w", tableName, err)
		return result
	}

	_, err = bm.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bm.bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		result.Error = fmt.Errorf("failed to upload backup for table %s to S3: %w", tableName, err)
		return result
	}

	bm.logger.Info("backup completed",
		"table", tableName,
		"items", len(items),
		"key", key,
	)

	return result
}

func (bm *BackupManager) scanTable(ctx context.Context, tableName string) ([]map[string]types.AttributeValue, error) {
	var allItems []map[string]types.AttributeValue
	var exclusiveStartKey map[string]types.AttributeValue

	for {
		limit := int32(maxScanLimit)
		input := &dynamodb.ScanInput{
			TableName:         aws.String(tableName),
			Limit:             &limit,
			ExclusiveStartKey: exclusiveStartKey,
		}

		output, err := bm.dynamoDBClient.Scan(ctx, input)
		if err != nil {
			return nil, err
		}

		allItems = append(allItems, output.Items...)

		if output.LastEvaluatedKey == nil {
			break
		}

		exclusiveStartKey = output.LastEvaluatedKey
	}

	return allItems, nil
}
