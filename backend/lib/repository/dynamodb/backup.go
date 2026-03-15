package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"golang.org/x/sync/errgroup"
)

const (
	maxScanLimit = 100
)

// DynamoDBScanner defines the interface for scanning DynamoDB tables.
type DynamoDBScanner interface {
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

// S3Putter defines the interface for putting objects to S3.
type S3Putter interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

// BackupResult represents the result of a backup operation.
type BackupResult struct {
	TableName  string
	ItemsCount int
	Key        string
	Error      error
	Latency    time.Duration
}

// BackupRepository handles backup operations for DynamoDB tables.
type BackupRepository struct {
	dynamoClient DynamoDBScanner
	s3Client     S3Putter
	bucket       string
	logger       *slog.Logger
}

// NewBackupRepository creates a new BackupRepository instance from the given configuration.
func NewBackupRepository(cfg *config.Config) *BackupRepository {
	awsConfig, _ := awsconfig.LoadDefaultConfig(context.TODO())

	return &BackupRepository{
		dynamoClient: dynamodb.NewFromConfig(awsConfig),
		s3Client:     s3.NewFromConfig(awsConfig),
		bucket:       cfg.BackupBucketName,
		logger:       slog.Default(),
	}
}

// BackupAllTables backs up all specified DynamoDB tables.
func (b *BackupRepository) BackupAllTables(ctx context.Context, tables []string) []*BackupResult {
	results := make([]*BackupResult, len(tables))
	eg := errgroup.Group{}

	for i, tableName := range tables {
		eg.Go(func() error {
			results[i] = b.BackupTable(ctx, tableName)
			return nil
		})
	}

	_ = eg.Wait()

	return results
}

// BackupTable backs up a single DynamoDB table.
func (b *BackupRepository) BackupTable(ctx context.Context, tableName string) *BackupResult {
	start := time.Now()
	result := &BackupResult{
		TableName: tableName,
	}

	if b.bucket == "" {
		result.Error = errors.New("backup bucket name is not configured")
		result.Latency = time.Since(start)
		return result
	}

	items, err := b.scanTable(ctx, tableName)
	if err != nil {
		result.Error = fmt.Errorf("failed to scan table %s: %w", tableName, err)
		result.Latency = time.Since(start)
		return result
	}

	result.ItemsCount = len(items)

	key := fmt.Sprintf("backups/%s-%s.json", tableName, time.Now().UTC().Format("20060102-150405"))
	result.Key = key

	backupData := map[string]any{
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"table_name": tableName,
		"item_count": len(items),
		"items":      items,
	}

	data, err := json.MarshalIndent(backupData, "", "  ")
	if err != nil {
		result.Error = fmt.Errorf("failed to marshal backup data for table %s: %w", tableName, err)
		result.Latency = time.Since(start)
		return result
	}

	_, err = b.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		result.Error = fmt.Errorf("failed to upload backup for table %s to S3: %w", tableName, err)
		result.Latency = time.Since(start)
		return result
	}

	result.Latency = time.Since(start)

	return result
}

func (b *BackupRepository) scanTable(ctx context.Context, tableName string) ([]map[string]types.AttributeValue, error) {
	var allItems []map[string]types.AttributeValue
	var exclusiveStartKey map[string]types.AttributeValue

	for {
		limit := int32(maxScanLimit)
		input := &dynamodb.ScanInput{
			TableName:         aws.String(tableName),
			Limit:             &limit,
			ExclusiveStartKey: exclusiveStartKey,
		}

		output, err := b.dynamoClient.Scan(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}

		allItems = append(allItems, output.Items...)

		if output.LastEvaluatedKey == nil {
			break
		}

		exclusiveStartKey = output.LastEvaluatedKey
	}

	return allItems, nil
}
