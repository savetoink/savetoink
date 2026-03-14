package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"golang.org/x/sync/errgroup"
)

const (
	maxScanLimit       = 100
	minBackupNameParts = 2
)

// DynamoDBScanner defines the interface for scanning DynamoDB tables.
type DynamoDBScanner interface {
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

// S3Putter defines the interface for putting objects to S3.
type S3Putter interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

// S3Getter defines the interface for getting objects from S3.
type S3Getter interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// DynamoDBBatchWriter defines the interface for batch writing to DynamoDB.
type DynamoDBBatchWriter interface {
	BatchWriteItem(
		ctx context.Context,
		params *dynamodb.BatchWriteItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.BatchWriteItemOutput, error)
}

// BackupResult represents the result of a backup operation.
type BackupResult struct {
	TableName  string
	ItemsCount int
	Key        string
	Error      error
	Latency    time.Duration
}

// RestoreResult represents the result of a restore operation.
type RestoreResult struct {
	TableName     string
	BackupName    string
	ItemsRestored int
	ItemsSkipped  int
	ItemsDeleted  int
	Error         error
	Latency       time.Duration
	Overwrite     bool
}

// BackupRepository handles backup operations for DynamoDB tables.
type BackupRepository struct {
	dynamoClient DynamoDBScanner
	s3Client     S3Putter
	s3Getter     S3Getter
	ddbWriter    DynamoDBBatchWriter
	bucket       string
	logger       *slog.Logger
}

// NewBackupRepository creates a new BackupRepository instance from the given configuration.
func NewBackupRepository(cfg *config.Config) *BackupRepository {
	awsConfig, _ := awsconfig.LoadDefaultConfig(context.TODO())
	ddbClient := dynamodb.NewFromConfig(awsConfig)
	s3Client := s3.NewFromConfig(awsConfig)

	return &BackupRepository{
		dynamoClient: ddbClient,
		s3Client:     s3Client,
		s3Getter:     s3Client,
		ddbWriter:    ddbClient,
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

	key, err := b.uploadBackup(ctx, tableName, items)
	if err != nil {
		result.Error = err
		result.Latency = time.Since(start)
		return result
	}
	result.Key = key
	result.Latency = time.Since(start)

	return result
}

func (b *BackupRepository) uploadBackup(
	ctx context.Context,
	tableName string,
	items []map[string]types.AttributeValue,
) (string, error) {
	key := fmt.Sprintf("backups/%s-%s.json", tableName, time.Now().UTC().Format("20060102-150405"))

	itemsArray := make([]json.RawMessage, 0, len(items))
	for _, item := range items {
		itemJSON, marshalErr := attributevalue.MarshalMapJSON(item)
		if marshalErr != nil {
			return "", fmt.Errorf("failed to marshal item: %w", marshalErr)
		}
		itemsArray = append(itemsArray, json.RawMessage(itemJSON))
	}

	backupData := map[string]any{
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"table_name": tableName,
		"item_count": len(items),
		"items":      itemsArray,
	}

	data, err := json.MarshalIndent(backupData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal backup data for table %s: %w", tableName, err)
	}

	_, err = b.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload backup for table %s to S3: %w", tableName, err)
	}

	return key, nil
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

// RestoreTable restores a DynamoDB table from a backup file in S3.
// If overwrite is true, the table will be cleared before restoring.
// Otherwise, existing items will be preserved.
func (b *BackupRepository) RestoreTable(
	ctx context.Context,
	backupName string,
	overwrite bool,
) *RestoreResult {
	start := time.Now()
	result := &RestoreResult{
		BackupName: backupName,
		Overwrite:  overwrite,
	}

	if b.bucket == "" {
		result.setError(time.Since(start), "backup bucket name is not configured")
		return result
	}

	key := "backups/" + backupName

	tableName, extractErr := b.extractTableName(backupName)
	if extractErr != nil {
		result.setError(time.Since(start), fmt.Sprintf("invalid backup name %s: %v", backupName, extractErr))
		return result
	}
	result.TableName = tableName

	backupData, downloadErr := b.downloadBackup(ctx, key)
	if downloadErr != nil {
		result.setError(time.Since(start), fmt.Sprintf("failed to download backup %s: %v", backupName, downloadErr))
		return result
	}

	items, parseErr := b.parseBackupData(backupData)
	if parseErr != nil {
		result.setError(time.Since(start), fmt.Sprintf("failed to parse backup data: %v", parseErr))
		return result
	}

	if overwrite {
		deleted, clearErr := b.clearTable(ctx, tableName)
		if clearErr != nil {
			result.setError(time.Since(start), fmt.Sprintf("failed to clear table %s: %v", tableName, clearErr))
			return result
		}
		result.ItemsDeleted = deleted
	} else {
		existingKeys := b.fetchExistingKeys(ctx, tableName)
		items = b.filterExistingItems(items, existingKeys)
		result.ItemsSkipped = len(existingKeys)
	}

	restored, writeErr := b.writeItems(ctx, tableName, items)
	if writeErr != nil {
		result.setError(time.Since(start), fmt.Sprintf("failed to restore items to table %s: %v", tableName, writeErr))
		return result
	}
	result.ItemsRestored = restored

	result.Latency = time.Since(start)
	return result
}

func (r *RestoreResult) setError(latency time.Duration, msg string) {
	r.Error = errors.New(msg)
	r.Latency = latency
}

func (b *BackupRepository) extractTableName(backupName string) (string, error) {
	if !strings.HasSuffix(backupName, ".json") {
		return "", errors.New("backup name must end with .json")
	}

	parts := strings.Split(strings.TrimSuffix(backupName, ".json"), "-")
	if len(parts) < minBackupNameParts {
		return "", errors.New("backup name format should be: tablename-timestamp.json")
	}

	tableNamePrefix := strings.Join(parts[:len(parts)-2], "-")
	tableName := strings.TrimPrefix(tableNamePrefix, "savetoink-")
	return tableName, nil
}

func (b *BackupRepository) downloadBackup(ctx context.Context, key string) (map[string]any, error) {
	resp, err := b.s3Getter.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	data, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read backup data: %w", readErr)
	}

	var backupData map[string]any
	if unmarshalErr := json.Unmarshal(data, &backupData); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal backup data: %w", unmarshalErr)
	}

	return backupData, nil
}

func (b *BackupRepository) parseBackupData(backupData map[string]any) ([]map[string]types.AttributeValue, error) {
	itemsRaw, hasItems := backupData["items"]
	if !hasItems {
		return nil, errors.New("backup data missing 'items' field")
	}

	items := make([]map[string]types.AttributeValue, 0)
	if itemsArray, isArray := itemsRaw.([]any); isArray {
		for _, itemRaw := range itemsArray {
			itemJSON, _ := json.Marshal(itemRaw)
			itemMap, err := attributevalue.UnmarshalMapJSON(itemJSON)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal item: %w", err)
			}
			items = append(items, itemMap)
		}
	}

	return items, nil
}

func (b *BackupRepository) clearTable(ctx context.Context, tableName string) (int, error) {
	items, err := b.scanTable(ctx, tableName)
	if err != nil {
		return 0, fmt.Errorf("failed to scan table: %w", err)
	}

	if len(items) == 0 {
		return 0, nil
	}

	deleted := 0
	batchSize := 25

	for i := 0; i < len(items); i += batchSize {
		end := min(i+batchSize, len(items))

		writeReqs := make([]types.WriteRequest, 0, end-i)
		for j := i; j < end; j++ {
			writeReqs = append(writeReqs, types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: items[j],
				},
			})
		}

		_, batchErr := b.ddbWriter.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				tableName: writeReqs,
			},
		})
		if batchErr != nil {
			return deleted, fmt.Errorf("failed to delete batch: %w", batchErr)
		}

		deleted += len(writeReqs)
	}

	return deleted, nil
}

func (b *BackupRepository) fetchExistingKeys(ctx context.Context, tableName string) map[string]struct{} {
	items, err := b.scanTable(ctx, tableName)
	if err != nil {
		return make(map[string]struct{})
	}

	keys := make(map[string]struct{})
	for _, item := range items {
		key := b.getItemKey(item)
		if key != "" {
			keys[key] = struct{}{}
		}
	}
	return keys
}

func (b *BackupRepository) filterExistingItems(
	items []map[string]types.AttributeValue,
	existingKeys map[string]struct{},
) []map[string]types.AttributeValue {
	filtered := make([]map[string]types.AttributeValue, 0, len(items))
	for _, item := range items {
		key := b.getItemKey(item)
		if key == "" {
			filtered = append(filtered, item)
			continue
		}
		if _, exists := existingKeys[key]; !exists {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func (b *BackupRepository) getItemKey(item map[string]types.AttributeValue) string {
	account, hasAccount := item["account"]
	if !hasAccount {
		return ""
	}

	keyBuilder := strings.Builder{}

	if accMember, isAccString := account.(*types.AttributeValueMemberS); isAccString {
		keyBuilder.WriteString(accMember.Value)
	}
	keyBuilder.WriteString(":")

	id, hasID := item["id"]
	if hasID {
		if idMember, isIDString := id.(*types.AttributeValueMemberS); isIDString {
			keyBuilder.WriteString(idMember.Value)
		}
	}

	return keyBuilder.String()
}

func (b *BackupRepository) writeItems(
	ctx context.Context,
	tableName string,
	items []map[string]types.AttributeValue,
) (int, error) {
	if len(items) == 0 {
		return 0, nil
	}

	batchSize := 25
	written := 0

	for i := 0; i < len(items); i += batchSize {
		end := min(i+batchSize, len(items))

		writeReqs := make([]types.WriteRequest, 0, end-i)
		for j := i; j < end; j++ {
			writeReqs = append(writeReqs, types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: items[j],
				},
			})
		}

		_, err := b.ddbWriter.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				tableName: writeReqs,
			},
		})
		if err != nil {
			return written, fmt.Errorf("failed to write batch: %w", err)
		}

		written += len(writeReqs)
	}

	return written, nil
}
