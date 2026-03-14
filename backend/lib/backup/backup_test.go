package backup

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDynamoDBScanner struct {
	scanFunc func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

func (m *mockDynamoDBScanner) Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	return m.scanFunc(ctx, params, optFns...)
}

type mockS3Putter struct {
	putObjectFunc func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

func (m *mockS3Putter) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return m.putObjectFunc(ctx, params, optFns...)
}

func TestBackupManager_BackupTable(t *testing.T) {
	t.Run("successful backup", func(t *testing.T) {
		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{
						{"id": &types.AttributeValueMemberS{Value: "item1"}},
						{"id": &types.AttributeValueMemberS{Value: "item2"}},
					},
				}, nil
			},
		}

		mockS3 := &mockS3Putter{
			putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				return &s3.PutObjectOutput{}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		bm := NewBackupManager(mockDDB, mockS3, "test-bucket", logger)

		result := bm.BackupTable(context.Background(), "test-table")

		require.NoError(t, result.Error)
		assert.Equal(t, "test-table", result.TableName)
		assert.Equal(t, 2, result.ItemsCount)
		assert.Contains(t, result.Key, "backups/test-table-")
	})

	t.Run("backup with empty bucket name", func(t *testing.T) {
		mockDDB := &mockDynamoDBScanner{}
		mockS3 := &mockS3Putter{}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		bm := NewBackupManager(mockDDB, mockS3, "", logger)

		result := bm.BackupTable(context.Background(), "test-table")

		require.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "backup bucket name is not configured")
	})

	t.Run("backup with scan error", func(t *testing.T) {
		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
				return nil, errors.New("scan error")
			},
		}

		mockS3 := &mockS3Putter{}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		bm := NewBackupManager(mockDDB, mockS3, "test-bucket", logger)

		result := bm.BackupTable(context.Background(), "test-table")

		require.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "failed to scan table")
	})

	t.Run("backup with s3 upload error", func(t *testing.T) {
		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{
						{"id": &types.AttributeValueMemberS{Value: "item1"}},
					},
				}, nil
			},
		}

		mockS3 := &mockS3Putter{
			putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				return nil, errors.New("s3 upload error")
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		bm := NewBackupManager(mockDDB, mockS3, "test-bucket", logger)

		result := bm.BackupTable(context.Background(), "test-table")

		require.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "failed to upload backup")
	})

	t.Run("backup with pagination", func(t *testing.T) {
		callCount := 0
		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
				callCount++
				if callCount == 1 {
					return &dynamodb.ScanOutput{
						Items: []map[string]types.AttributeValue{
							{"id": &types.AttributeValueMemberS{Value: "item1"}},
						},
						LastEvaluatedKey: map[string]types.AttributeValue{
							"id": &types.AttributeValueMemberS{Value: "key1"},
						},
					}, nil
				}
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{
						{"id": &types.AttributeValueMemberS{Value: "item2"}},
					},
				}, nil
			},
		}

		mockS3 := &mockS3Putter{
			putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				return &s3.PutObjectOutput{}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		bm := NewBackupManager(mockDDB, mockS3, "test-bucket", logger)

		result := bm.BackupTable(context.Background(), "test-table")

		require.NoError(t, result.Error)
		assert.Equal(t, 2, result.ItemsCount)
		assert.Equal(t, 2, callCount)
	})
}

func TestBackupManager_BackupAllTables(t *testing.T) {
	t.Run("successful backup of multiple tables", func(t *testing.T) {
		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{
						{"id": &types.AttributeValueMemberS{Value: "item1"}},
					},
				}, nil
			},
		}

		mockS3 := &mockS3Putter{
			putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				return &s3.PutObjectOutput{}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		bm := NewBackupManager(mockDDB, mockS3, "test-bucket", logger)

		tables := []string{"table1", "table2", "table3"}
		results := bm.BackupAllTables(context.Background(), tables)

		assert.Len(t, results, 3)
		for _, result := range results {
			assert.NoError(t, result.Error)
			assert.Equal(t, 1, result.ItemsCount)
		}
	})

	t.Run("partial failure", func(t *testing.T) {
		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
				if *params.TableName == "table2" {
					return nil, errors.New("scan error")
				}
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{
						{"id": &types.AttributeValueMemberS{Value: "item1"}},
					},
				}, nil
			},
		}

		mockS3 := &mockS3Putter{
			putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				return &s3.PutObjectOutput{}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		bm := NewBackupManager(mockDDB, mockS3, "test-bucket", logger)

		tables := []string{"table1", "table2", "table3"}
		results := bm.BackupAllTables(context.Background(), tables)

		assert.Len(t, results, 3)
		assert.NoError(t, results[0].Error)
		assert.Error(t, results[1].Error)
		assert.NoError(t, results[2].Error)
	})
}

func TestBackupManager_ScanTable(t *testing.T) {
	t.Run("empty table", func(t *testing.T) {
		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{},
				}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		bm := NewBackupManager(mockDDB, nil, "test-bucket", logger)

		items, err := bm.scanTable(context.Background(), "test-table")

		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("large table requiring pagination", func(t *testing.T) {
		itemsPerScan := 50
		totalItems := 250
		scanCount := 0

		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
				scanCount++
				items := make([]map[string]types.AttributeValue, itemsPerScan)
				for i := 0; i < itemsPerScan; i++ {
					items[i] = map[string]types.AttributeValue{
						"id": &types.AttributeValueMemberS{Value: "item"},
					}
				}

				if scanCount*itemsPerScan >= totalItems {
					return &dynamodb.ScanOutput{
						Items: items[:totalItems-(scanCount-1)*itemsPerScan],
					}, nil
				}

				return &dynamodb.ScanOutput{
					Items: items,
					LastEvaluatedKey: map[string]types.AttributeValue{
						"id": &types.AttributeValueMemberS{Value: "key"},
					},
				}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		bm := NewBackupManager(mockDDB, nil, "test-bucket", logger)

		items, err := bm.scanTable(context.Background(), "test-table")

		require.NoError(t, err)
		assert.Equal(t, totalItems, len(items))
	})
}

func TestBackupManager_KeyGeneration(t *testing.T) {
	mockDDB := &mockDynamoDBScanner{
		scanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			return &dynamodb.ScanOutput{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "item1"}},
				},
			}, nil
		},
	}

	mockS3 := &mockS3Putter{
		putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return &s3.PutObjectOutput{}, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	bm := NewBackupManager(mockDDB, mockS3, "test-bucket", logger)

	result := bm.BackupTable(context.Background(), "test-table")

	require.NoError(t, result.Error)
	assert.Contains(t, result.Key, "backups/test-table-")
	assert.Contains(t, result.Key, ".json")
}

func TestBackupManager_TimestampValidation(t *testing.T) {
	mockDDB := &mockDynamoDBScanner{
		scanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			return &dynamodb.ScanOutput{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "item1"}},
				},
			}, nil
		},
	}

	mockS3 := &mockS3Putter{
		putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return &s3.PutObjectOutput{}, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	bm := NewBackupManager(mockDDB, mockS3, "test-bucket", logger)

	result := bm.BackupTable(context.Background(), "test-table")

	require.NoError(t, result.Error)
	require.Contains(t, result.Key, "backups/test-table-")
	require.Contains(t, result.Key, ".json")

	parsedTime, err := time.Parse("20060102-150405", result.Key[len("backups/test-table-"):len(result.Key)-len(".json")])
	require.NoError(t, err)
	require.NotNil(t, parsedTime)
}
