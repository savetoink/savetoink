package repository

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDynamoDBScanner struct {
	scanFunc func(
		ctx context.Context,
		params *dynamodb.ScanInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.ScanOutput, error)
}

func (m *mockDynamoDBScanner) Scan(
	ctx context.Context,
	params *dynamodb.ScanInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.ScanOutput, error) {
	return m.scanFunc(ctx, params, optFns...)
}

type mockS3Putter struct {
	putObjectFunc func(
		ctx context.Context,
		params *s3.PutObjectInput,
		optFns ...func(*s3.Options),
	) (*s3.PutObjectOutput, error)
}

func (m *mockS3Putter) PutObject(
	ctx context.Context,
	params *s3.PutObjectInput,
	optFns ...func(*s3.Options),
) (*s3.PutObjectOutput, error) {
	return m.putObjectFunc(ctx, params, optFns...)
}

type mockS3Getter struct {
	getObjectFunc func(
		ctx context.Context,
		params *s3.GetObjectInput,
		optFns ...func(*s3.Options),
	) (*s3.GetObjectOutput, error)
}

func (m *mockS3Getter) GetObject(
	ctx context.Context,
	params *s3.GetObjectInput,
	optFns ...func(*s3.Options),
) (*s3.GetObjectOutput, error) {
	return m.getObjectFunc(ctx, params, optFns...)
}

type mockDynamoDBBatchWriter struct {
	batchWriteItemFunc func(
		ctx context.Context,
		params *dynamodb.BatchWriteItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.BatchWriteItemOutput, error)
}

type mockDynamoDBDescriber struct {
	describeTableFunc func(
		ctx context.Context,
		params *dynamodb.DescribeTableInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.DescribeTableOutput, error)
}

func (m *mockDynamoDBDescriber) DescribeTable(
	ctx context.Context,
	params *dynamodb.DescribeTableInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.DescribeTableOutput, error) {
	return m.describeTableFunc(ctx, params, optFns...)
}

func (m *mockDynamoDBBatchWriter) BatchWriteItem(
	ctx context.Context,
	params *dynamodb.BatchWriteItemInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.BatchWriteItemOutput, error) {
	return m.batchWriteItemFunc(ctx, params, optFns...)
}

func TestBackupRepository_BackupTable(t *testing.T) {
	t.Run("successful backup", func(t *testing.T) {
		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(
				_ context.Context,
				_ *dynamodb.ScanInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.ScanOutput, error) {
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{
						{"id": &types.AttributeValueMemberS{Value: "item1"}},
						{"id": &types.AttributeValueMemberS{Value: "item2"}},
					},
				}, nil
			},
		}

		mockDescriber := &mockDynamoDBDescriber{
			describeTableFunc: func(
				_ context.Context,
				_ *dynamodb.DescribeTableInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.DescribeTableOutput, error) {
				return &dynamodb.DescribeTableOutput{
					Table: &types.TableDescription{
						KeySchema: []types.KeySchemaElement{
							{AttributeName: aws.String("account"), KeyType: types.KeyTypeHash},
							{AttributeName: aws.String("id"), KeyType: types.KeyTypeRange},
						},
					},
				}, nil
			},
		}

		mockS3 := &mockS3Putter{
			putObjectFunc: func(
				_ context.Context,
				_ *s3.PutObjectInput,
				_ ...func(*s3.Options),
			) (*s3.PutObjectOutput, error) {
				return &s3.PutObjectOutput{}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			dynamoClient: mockDDB,
			ddbDescriber: mockDescriber,
			s3Client:     mockS3,
			bucket:       "test-bucket",
			logger:       logger,
		}

		result := br.BackupTable(context.Background(), "test-table")

		require.NoError(t, result.Error)
		assert.Equal(t, "test-table", result.TableName)
		assert.Equal(t, 2, result.ItemsCount)
		assert.Contains(t, result.Key, "backups/test-table-")
	})

	t.Run("backup with empty bucket name", func(t *testing.T) {
		mockDDB := &mockDynamoDBScanner{}
		mockS3 := &mockS3Putter{}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			dynamoClient: mockDDB,
			s3Client:     mockS3,
			bucket:       "",
			logger:       logger,
		}

		result := br.BackupTable(context.Background(), "test-table")

		require.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "backup bucket name is not configured")
	})

	t.Run("backup with scan error", func(t *testing.T) {
		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(
				_ context.Context,
				_ *dynamodb.ScanInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.ScanOutput, error) {
				return nil, errors.New("scan error")
			},
		}

		mockS3 := &mockS3Putter{}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			dynamoClient: mockDDB,
			s3Client:     mockS3,
			bucket:       "test-bucket",
			logger:       logger,
		}

		result := br.BackupTable(context.Background(), "test-table")

		require.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "failed to scan table")
	})

	t.Run("backup with s3 upload error", func(t *testing.T) {
		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(
				_ context.Context,
				_ *dynamodb.ScanInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.ScanOutput, error) {
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{
						{"id": &types.AttributeValueMemberS{Value: "item1"}},
					},
				}, nil
			},
		}

		mockDescriber := &mockDynamoDBDescriber{
			describeTableFunc: func(
				_ context.Context,
				_ *dynamodb.DescribeTableInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.DescribeTableOutput, error) {
				return &dynamodb.DescribeTableOutput{
					Table: &types.TableDescription{
						KeySchema: []types.KeySchemaElement{
							{AttributeName: aws.String("account"), KeyType: types.KeyTypeHash},
							{AttributeName: aws.String("id"), KeyType: types.KeyTypeRange},
						},
					},
				}, nil
			},
		}

		mockS3 := &mockS3Putter{
			putObjectFunc: func(
				_ context.Context,
				_ *s3.PutObjectInput,
				_ ...func(*s3.Options),
			) (*s3.PutObjectOutput, error) {
				return nil, errors.New("s3 upload error")
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			dynamoClient: mockDDB,
			ddbDescriber: mockDescriber,
			s3Client:     mockS3,
			bucket:       "test-bucket",
			logger:       logger,
		}

		result := br.BackupTable(context.Background(), "test-table")

		require.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "failed to upload backup")
	})

	t.Run("backup with pagination", func(t *testing.T) {
		callCount := 0
		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(
				_ context.Context,
				_ *dynamodb.ScanInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.ScanOutput, error) {
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

		mockDescriber := &mockDynamoDBDescriber{
			describeTableFunc: func(
				_ context.Context,
				_ *dynamodb.DescribeTableInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.DescribeTableOutput, error) {
				return &dynamodb.DescribeTableOutput{
					Table: &types.TableDescription{
						KeySchema: []types.KeySchemaElement{
							{AttributeName: aws.String("account"), KeyType: types.KeyTypeHash},
							{AttributeName: aws.String("id"), KeyType: types.KeyTypeRange},
						},
					},
				}, nil
			},
		}

		mockS3 := &mockS3Putter{
			putObjectFunc: func(
				_ context.Context,
				_ *s3.PutObjectInput,
				_ ...func(*s3.Options),
			) (*s3.PutObjectOutput, error) {
				return &s3.PutObjectOutput{}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			dynamoClient: mockDDB,
			ddbDescriber: mockDescriber,
			s3Client:     mockS3,
			bucket:       "test-bucket",
			logger:       logger,
		}

		result := br.BackupTable(context.Background(), "test-table")

		require.NoError(t, result.Error)
		assert.Equal(t, 2, result.ItemsCount)
		assert.Equal(t, 2, callCount)
	})
}

func TestBackupRepository_BackupAllTables(t *testing.T) {
	t.Run("successful backup of multiple tables", func(t *testing.T) {
		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(
				_ context.Context,
				_ *dynamodb.ScanInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.ScanOutput, error) {
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{
						{"id": &types.AttributeValueMemberS{Value: "item1"}},
					},
				}, nil
			},
		}

		mockDescriber := &mockDynamoDBDescriber{
			describeTableFunc: func(
				_ context.Context,
				_ *dynamodb.DescribeTableInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.DescribeTableOutput, error) {
				return &dynamodb.DescribeTableOutput{
					Table: &types.TableDescription{
						KeySchema: []types.KeySchemaElement{
							{AttributeName: aws.String("account"), KeyType: types.KeyTypeHash},
							{AttributeName: aws.String("id"), KeyType: types.KeyTypeRange},
						},
					},
				}, nil
			},
		}

		mockS3 := &mockS3Putter{
			putObjectFunc: func(
				_ context.Context,
				_ *s3.PutObjectInput,
				_ ...func(*s3.Options),
			) (*s3.PutObjectOutput, error) {
				return &s3.PutObjectOutput{}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			dynamoClient: mockDDB,
			ddbDescriber: mockDescriber,
			s3Client:     mockS3,
			bucket:       "test-bucket",
			logger:       logger,
		}

		tables := []string{"table1", "table2", "table3"}
		results := br.BackupAllTables(context.Background(), tables)

		assert.Len(t, results, 3)
		for _, result := range results {
			assert.NoError(t, result.Error)
			assert.Equal(t, 1, result.ItemsCount)
		}
	})

	t.Run("partial failure", func(t *testing.T) {
		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(
				_ context.Context,
				params *dynamodb.ScanInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.ScanOutput, error) {
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

		mockDescriber := &mockDynamoDBDescriber{
			describeTableFunc: func(
				_ context.Context,
				_ *dynamodb.DescribeTableInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.DescribeTableOutput, error) {
				return &dynamodb.DescribeTableOutput{
					Table: &types.TableDescription{
						KeySchema: []types.KeySchemaElement{
							{AttributeName: aws.String("account"), KeyType: types.KeyTypeHash},
							{AttributeName: aws.String("id"), KeyType: types.KeyTypeRange},
						},
					},
				}, nil
			},
		}

		mockS3 := &mockS3Putter{
			putObjectFunc: func(
				_ context.Context,
				_ *s3.PutObjectInput,
				_ ...func(*s3.Options),
			) (*s3.PutObjectOutput, error) {
				return &s3.PutObjectOutput{}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			dynamoClient: mockDDB,
			ddbDescriber: mockDescriber,
			s3Client:     mockS3,
			bucket:       "test-bucket",
			logger:       logger,
		}

		tables := []string{"table1", "table2", "table3"}
		results := br.BackupAllTables(context.Background(), tables)

		assert.Len(t, results, 3)
		assert.NoError(t, results[0].Error)
		assert.Error(t, results[1].Error)
		assert.NoError(t, results[2].Error)
	})
}

func TestBackupRepository_ScanTable(t *testing.T) {
	t.Run("empty table", func(t *testing.T) {
		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(
				_ context.Context,
				_ *dynamodb.ScanInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.ScanOutput, error) {
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{},
				}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			dynamoClient: mockDDB,
			logger:       logger,
		}

		items, err := br.scanTable(context.Background(), "test-table")

		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("large table requiring pagination", func(t *testing.T) {
		itemsPerScan := 50
		totalItems := 250
		scanCount := 0

		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(
				_ context.Context,
				_ *dynamodb.ScanInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.ScanOutput, error) {
				scanCount++
				items := make([]map[string]types.AttributeValue, itemsPerScan)
				for i := range itemsPerScan {
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
		br := &BackupRepository{
			dynamoClient: mockDDB,
			logger:       logger,
		}

		items, err := br.scanTable(context.Background(), "test-table")

		require.NoError(t, err)
		assert.Equal(t, totalItems, len(items))
	})
}

func TestBackupRepository_KeyGeneration(t *testing.T) {
	mockDDB := &mockDynamoDBScanner{
		scanFunc: func(
			_ context.Context,
			_ *dynamodb.ScanInput,
			_ ...func(*dynamodb.Options),
		) (*dynamodb.ScanOutput, error) {
			return &dynamodb.ScanOutput{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "item1"}},
				},
			}, nil
		},
	}

	mockDescriber := &mockDynamoDBDescriber{
		describeTableFunc: func(
			_ context.Context,
			_ *dynamodb.DescribeTableInput,
			_ ...func(*dynamodb.Options),
		) (*dynamodb.DescribeTableOutput, error) {
			return &dynamodb.DescribeTableOutput{
				Table: &types.TableDescription{
					KeySchema: []types.KeySchemaElement{
						{AttributeName: aws.String("account"), KeyType: types.KeyTypeHash},
						{AttributeName: aws.String("id"), KeyType: types.KeyTypeRange},
					},
				},
			}, nil
		},
	}

	mockS3 := &mockS3Putter{
		putObjectFunc: func(
			_ context.Context,
			_ *s3.PutObjectInput,
			_ ...func(*s3.Options),
		) (*s3.PutObjectOutput, error) {
			return &s3.PutObjectOutput{}, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	br := &BackupRepository{
		dynamoClient: mockDDB,
		ddbDescriber: mockDescriber,
		s3Client:     mockS3,
		bucket:       "test-bucket",
		logger:       logger,
	}

	result := br.BackupTable(context.Background(), "test-table")

	require.NoError(t, result.Error)
	assert.Contains(t, result.Key, "backups/test-table-")
	assert.Contains(t, result.Key, ".json.gz")
}

func TestBackupRepository_TimestampValidation(t *testing.T) {
	mockDDB := &mockDynamoDBScanner{
		scanFunc: func(
			_ context.Context,
			_ *dynamodb.ScanInput,
			_ ...func(*dynamodb.Options),
		) (*dynamodb.ScanOutput, error) {
			return &dynamodb.ScanOutput{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "item1"}},
				},
			}, nil
		},
	}

	mockDescriber := &mockDynamoDBDescriber{
		describeTableFunc: func(
			_ context.Context,
			_ *dynamodb.DescribeTableInput,
			_ ...func(*dynamodb.Options),
		) (*dynamodb.DescribeTableOutput, error) {
			return &dynamodb.DescribeTableOutput{
				Table: &types.TableDescription{
					KeySchema: []types.KeySchemaElement{
						{AttributeName: aws.String("account"), KeyType: types.KeyTypeHash},
						{AttributeName: aws.String("id"), KeyType: types.KeyTypeRange},
					},
				},
			}, nil
		},
	}

	mockS3 := &mockS3Putter{
		putObjectFunc: func(
			_ context.Context,
			_ *s3.PutObjectInput,
			_ ...func(*s3.Options),
		) (*s3.PutObjectOutput, error) {
			return &s3.PutObjectOutput{}, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	br := &BackupRepository{
		dynamoClient: mockDDB,
		ddbDescriber: mockDescriber,
		s3Client:     mockS3,
		bucket:       "test-bucket",
		logger:       logger,
	}

	result := br.BackupTable(context.Background(), "test-table")

	require.NoError(t, result.Error)
	require.Contains(t, result.Key, "backups/test-table-")
	require.Contains(t, result.Key, ".json.gz")

	timestampStart := len("backups/test-table-")
	timestampEnd := len(result.Key) - len(".json.gz")
	timestamp := result.Key[timestampStart:timestampEnd]
	parsedTime, err := time.Parse("20060102-150405Z", timestamp)
	require.NoError(t, err)
	require.NotNil(t, parsedTime)
}

const (
	testBackupJSON = `{
		"timestamp": "2024-03-15T13:13:17Z",
		"table_name": "test-articles",
		"key_schema": {"hash_key": "account", "range_key": "id"},
		"item_count": 2,
		"items": [
			{"account": {"S": "account1"}, "id": {"S": "id1"}},
			{"account": {"S": "account1"}, "id": {"S": "id2"}}
		]
	}`

	testSingleItemBackupJSON = `{
		"timestamp": "2024-03-15T13:13:17Z",
		"table_name": "test-articles",
		"key_schema": {"hash_key": "account", "range_key": "id"},
		"item_count": 1,
		"items": [
			{"account": {"S": "account1"}, "id": {"S": "id1"}}
		]
	}`
)

func TestBackupRepository_RestoreTable_Overwrite(t *testing.T) {
	t.Run("successful restore with overwrite", func(t *testing.T) {
		backupJSON := testBackupJSON

		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(
				_ context.Context,
				_ *dynamodb.ScanInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.ScanOutput, error) {
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{
						{
							"account": &types.AttributeValueMemberS{Value: "account1"},
							"id":      &types.AttributeValueMemberS{Value: "old-id"},
						},
					},
				}, nil
			},
		}

		mockS3Getter := &mockS3Getter{
			getObjectFunc: func(
				_ context.Context,
				_ *s3.GetObjectInput,
				_ ...func(*s3.Options),
			) (*s3.GetObjectOutput, error) {
				var buf bytes.Buffer
				gzWriter := gzip.NewWriter(&buf)
				_, _ = gzWriter.Write([]byte(backupJSON))
				_ = gzWriter.Close()
				return &s3.GetObjectOutput{
					Body: io.NopCloser(bytes.NewReader(buf.Bytes())),
				}, nil
			},
		}

		mockDDBWriter := &mockDynamoDBBatchWriter{
			batchWriteItemFunc: func(
				_ context.Context,
				params *dynamodb.BatchWriteItemInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.BatchWriteItemOutput, error) {
				for _, writes := range params.RequestItems {
					for _, write := range writes {
						if write.PutRequest != nil {
							t.Log("Putting item:", write.PutRequest.Item)
						}
						if write.DeleteRequest != nil {
							t.Log("Deleting item:", write.DeleteRequest.Key)
						}
					}
				}
				return &dynamodb.BatchWriteItemOutput{}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			dynamoClient: mockDDB,
			s3Getter:     mockS3Getter,
			ddbWriter:    mockDDBWriter,
			bucket:       "test-bucket",
			logger:       logger,
		}

		result := br.RestoreTable(context.Background(), "articles-20240315-131317Z.json.gz", true)

		require.NoError(t, result.Error)
		assert.Equal(t, "articles", result.TableName)
		assert.Equal(t, "articles-20240315-131317Z.json.gz", result.BackupName)
		assert.Equal(t, 2, result.ItemsRestored)
		assert.Equal(t, 1, result.ItemsDeleted)
		assert.True(t, result.Overwrite)
		assert.Greater(t, result.Latency, time.Duration(0))
	})

	t.Run("successful restore without overwrite (merge)", func(t *testing.T) {
		backupJSON := testBackupJSON

		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(
				_ context.Context,
				_ *dynamodb.ScanInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.ScanOutput, error) {
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{
						{
							"account": &types.AttributeValueMemberS{Value: "account1"},
							"id":      &types.AttributeValueMemberS{Value: "id1"},
						},
					},
				}, nil
			},
		}

		mockS3Getter := &mockS3Getter{
			getObjectFunc: func(
				_ context.Context,
				_ *s3.GetObjectInput,
				_ ...func(*s3.Options),
			) (*s3.GetObjectOutput, error) {
				var buf bytes.Buffer
				gzWriter := gzip.NewWriter(&buf)
				_, _ = gzWriter.Write([]byte(backupJSON))
				_ = gzWriter.Close()
				return &s3.GetObjectOutput{
					Body: io.NopCloser(bytes.NewReader(buf.Bytes())),
				}, nil
			},
		}

		mockDDBWriter := &mockDynamoDBBatchWriter{
			batchWriteItemFunc: func(
				_ context.Context,
				params *dynamodb.BatchWriteItemInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.BatchWriteItemOutput, error) {
				for _, writes := range params.RequestItems {
					for _, write := range writes {
						if write.PutRequest != nil {
							t.Log("Putting item:", write.PutRequest.Item)
						}
					}
				}
				return &dynamodb.BatchWriteItemOutput{}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			dynamoClient: mockDDB,
			s3Getter:     mockS3Getter,
			ddbWriter:    mockDDBWriter,
			bucket:       "test-bucket",
			logger:       logger,
		}

		result := br.RestoreTable(context.Background(), "articles-20240315-131317Z.json.gz", false)

		require.NoError(t, result.Error)
		assert.Equal(t, "articles", result.TableName)
		assert.Equal(t, "articles-20240315-131317Z.json.gz", result.BackupName)
		assert.Equal(t, 1, result.ItemsRestored)
		assert.Equal(t, 1, result.ItemsSkipped)
		assert.Equal(t, 0, result.ItemsDeleted)
		assert.False(t, result.Overwrite)
	})

	t.Run("restore with invalid backup name format", func(t *testing.T) {
		mockS3Getter := &mockS3Getter{}
		mockDDBWriter := &mockDynamoDBBatchWriter{}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			s3Getter:  mockS3Getter,
			ddbWriter: mockDDBWriter,
			bucket:    "test-bucket",
			logger:    logger,
		}

		result := br.RestoreTable(context.Background(), "invalid-name", false)

		require.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "backup name format should be")
	})

	t.Run("restore with S3 error", func(t *testing.T) {
		mockS3Getter := &mockS3Getter{
			getObjectFunc: func(
				_ context.Context,
				_ *s3.GetObjectInput,
				_ ...func(*s3.Options),
			) (*s3.GetObjectOutput, error) {
				return nil, errors.New("s3 error")
			},
		}

		mockDDBWriter := &mockDynamoDBBatchWriter{}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			s3Getter:  mockS3Getter,
			ddbWriter: mockDDBWriter,
			bucket:    "test-bucket",
			logger:    logger,
		}

		result := br.RestoreTable(context.Background(), "articles-20240315-131317Z.json.gz", false)

		require.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "failed to download backup")
	})

	t.Run("restore with missing items field", func(t *testing.T) {
		backupJSON := `{
			"timestamp": "2024-03-15T13:13:17Z",
			"table_name": "test-articles"
		}`

		mockS3Getter := &mockS3Getter{
			getObjectFunc: func(
				_ context.Context,
				_ *s3.GetObjectInput,
				_ ...func(*s3.Options),
			) (*s3.GetObjectOutput, error) {
				var buf bytes.Buffer
				gzWriter := gzip.NewWriter(&buf)
				_, _ = gzWriter.Write([]byte(backupJSON))
				_ = gzWriter.Close()
				return &s3.GetObjectOutput{
					Body: io.NopCloser(bytes.NewReader(buf.Bytes())),
				}, nil
			},
		}

		mockDDBWriter := &mockDynamoDBBatchWriter{}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			s3Getter:  mockS3Getter,
			ddbWriter: mockDDBWriter,
			bucket:    "test-bucket",
			logger:    logger,
		}

		result := br.RestoreTable(context.Background(), "articles-20240315-131317Z.json.gz", false)

		require.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "backup data missing 'items' field")
	})

	t.Run("restore with DynamoDB batch write error", func(t *testing.T) {
		backupJSON := testSingleItemBackupJSON

		mockDDB := &mockDynamoDBScanner{
			scanFunc: func(
				_ context.Context,
				_ *dynamodb.ScanInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.ScanOutput, error) {
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{},
				}, nil
			},
		}

		mockS3Getter := &mockS3Getter{
			getObjectFunc: func(
				_ context.Context,
				_ *s3.GetObjectInput,
				_ ...func(*s3.Options),
			) (*s3.GetObjectOutput, error) {
				var buf bytes.Buffer
				gzWriter := gzip.NewWriter(&buf)
				_, _ = gzWriter.Write([]byte(backupJSON))
				_ = gzWriter.Close()
				return &s3.GetObjectOutput{
					Body: io.NopCloser(bytes.NewReader(buf.Bytes())),
				}, nil
			},
		}

		mockDDBWriter := &mockDynamoDBBatchWriter{
			batchWriteItemFunc: func(
				_ context.Context,
				_ *dynamodb.BatchWriteItemInput,
				_ ...func(*dynamodb.Options),
			) (*dynamodb.BatchWriteItemOutput, error) {
				return nil, errors.New("batch write error")
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			dynamoClient: mockDDB,
			s3Getter:     mockS3Getter,
			ddbWriter:    mockDDBWriter,
			bucket:       "test-bucket",
			logger:       logger,
		}

		result := br.RestoreTable(context.Background(), "articles-20240315-131317Z.json.gz", false)

		require.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "failed to restore items")
	})
}

func TestBackupRepository_ExtractTableName(t *testing.T) {
	t.Run("simple table name", func(t *testing.T) {
		mockS3Getter := &mockS3Getter{}
		mockDDBWriter := &mockDynamoDBBatchWriter{}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			s3Getter:  mockS3Getter,
			ddbWriter: mockDDBWriter,
			bucket:    "test-bucket",
			logger:    logger,
		}

		tableName, err := br.extractTableName("articles-20240315-131317Z.json.gz")
		require.NoError(t, err)
		assert.Equal(t, "articles", tableName)
	})

	t.Run("table name with prefix", func(t *testing.T) {
		mockS3Getter := &mockS3Getter{}
		mockDDBWriter := &mockDynamoDBBatchWriter{}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			s3Getter:  mockS3Getter,
			ddbWriter: mockDDBWriter,
			bucket:    "test-bucket",
			logger:    logger,
		}

		tableName, err := br.extractTableName("savetoink-articles-20240315-131317Z.json.gz")
		require.NoError(t, err)
		assert.Equal(t, "savetoink-articles", tableName)
	})

	t.Run("table name with multi prefix", func(t *testing.T) {
		mockS3Getter := &mockS3Getter{}
		mockDDBWriter := &mockDynamoDBBatchWriter{}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			s3Getter:  mockS3Getter,
			ddbWriter: mockDDBWriter,
			bucket:    "test-bucket",
			logger:    logger,
		}

		tableName, err := br.extractTableName("savetoink-multi-prefix-20240315-131317Z.json.gz")
		require.NoError(t, err)
		assert.Equal(t, "savetoink-multi-prefix", tableName)
	})

	t.Run("missing .json.gz extension", func(t *testing.T) {
		mockS3Getter := &mockS3Getter{}
		mockDDBWriter := &mockDynamoDBBatchWriter{}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			s3Getter:  mockS3Getter,
			ddbWriter: mockDDBWriter,
			bucket:    "test-bucket",
			logger:    logger,
		}

		_, err := br.extractTableName("articles-20240315-131317")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})

	t.Run("invalid format", func(t *testing.T) {
		mockS3Getter := &mockS3Getter{}
		mockDDBWriter := &mockDynamoDBBatchWriter{}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			s3Getter:  mockS3Getter,
			ddbWriter: mockDDBWriter,
			bucket:    "test-bucket",
			logger:    logger,
		}

		_, err := br.extractTableName("invalid.json.gz")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})

	t.Run("wrong date format with dashes", func(t *testing.T) {
		mockS3Getter := &mockS3Getter{}
		mockDDBWriter := &mockDynamoDBBatchWriter{}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			s3Getter:  mockS3Getter,
			ddbWriter: mockDDBWriter,
			bucket:    "test-bucket",
			logger:    logger,
		}

		_, err := br.extractTableName("articles-2024-03-15-131317.json.gz")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})

	t.Run("wrong time format with colons", func(t *testing.T) {
		mockS3Getter := &mockS3Getter{}
		mockDDBWriter := &mockDynamoDBBatchWriter{}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			s3Getter:  mockS3Getter,
			ddbWriter: mockDDBWriter,
			bucket:    "test-bucket",
			logger:    logger,
		}

		_, err := br.extractTableName("articles-20240315-13:13:17.json.gz")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})

	t.Run("missing time part", func(t *testing.T) {
		mockS3Getter := &mockS3Getter{}
		mockDDBWriter := &mockDynamoDBBatchWriter{}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			s3Getter:  mockS3Getter,
			ddbWriter: mockDDBWriter,
			bucket:    "test-bucket",
			logger:    logger,
		}

		_, err := br.extractTableName("articles-20240315.json.gz")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})

	t.Run("missing date part", func(t *testing.T) {
		mockS3Getter := &mockS3Getter{}
		mockDDBWriter := &mockDynamoDBBatchWriter{}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			s3Getter:  mockS3Getter,
			ddbWriter: mockDDBWriter,
			bucket:    "test-bucket",
			logger:    logger,
		}

		_, err := br.extractTableName("articles-131317.json.gz")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})

	t.Run("too many digits in date", func(t *testing.T) {
		mockS3Getter := &mockS3Getter{}
		mockDDBWriter := &mockDynamoDBBatchWriter{}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			s3Getter:  mockS3Getter,
			ddbWriter: mockDDBWriter,
			bucket:    "test-bucket",
			logger:    logger,
		}

		_, err := br.extractTableName("articles-202403151-131317Z.json.gz")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})

	t.Run("too few digits in time", func(t *testing.T) {
		mockS3Getter := &mockS3Getter{}
		mockDDBWriter := &mockDynamoDBBatchWriter{}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			s3Getter:  mockS3Getter,
			ddbWriter: mockDDBWriter,
			bucket:    "test-bucket",
			logger:    logger,
		}

		_, err := br.extractTableName("articles-20240315-1313.json.gz")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})

	t.Run("alpha characters in timestamp", func(t *testing.T) {
		mockS3Getter := &mockS3Getter{}
		mockDDBWriter := &mockDynamoDBBatchWriter{}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		br := &BackupRepository{
			s3Getter:  mockS3Getter,
			ddbWriter: mockDDBWriter,
			bucket:    "test-bucket",
			logger:    logger,
		}

		_, err := br.extractTableName("articles-abcdefghijklmnop-qrstuv.json.gz")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})
}

func TestGenerateBackupFilename(t *testing.T) {
	t.Run("generates correct filename", func(t *testing.T) {
		filename := generateBackupFilename("test-table")

		assert.Contains(t, filename, "backups/test-table-")
		assert.Contains(t, filename, ".json.gz")
		assert.Contains(t, filename, "Z")
	})

	t.Run("handles table name with prefix", func(t *testing.T) {
		filename := generateBackupFilename("savetoink-articles")

		assert.Contains(t, filename, "backups/savetoink-articles-")
		assert.Contains(t, filename, ".json.gz")
	})
}

func TestParseBackupFilename(t *testing.T) {
	t.Run("parses simple table name", func(t *testing.T) {
		tableName, err := ParseBackupFilename("articles-20240315-131317Z.json.gz")

		require.NoError(t, err)
		assert.Equal(t, "articles", tableName)
	})

	t.Run("parses table name with prefix", func(t *testing.T) {
		tableName, err := ParseBackupFilename("savetoink-articles-20240315-131317Z.json.gz")

		require.NoError(t, err)
		assert.Equal(t, "savetoink-articles", tableName)
	})

	t.Run("parses table name with multi prefix", func(t *testing.T) {
		tableName, err := ParseBackupFilename("savetoink-multi-prefix-20240315-131317Z.json.gz")

		require.NoError(t, err)
		assert.Equal(t, "savetoink-multi-prefix", tableName)
	})

	t.Run("missing .json.gz extension", func(t *testing.T) {
		_, err := ParseBackupFilename("articles-20240315-131317Z")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})

	t.Run("invalid format", func(t *testing.T) {
		_, err := ParseBackupFilename("invalid.json.gz")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})

	t.Run("wrong date format with dashes", func(t *testing.T) {
		_, err := ParseBackupFilename("articles-2024-03-15-131317Z.json.gz")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})

	t.Run("wrong time format with colons", func(t *testing.T) {
		_, err := ParseBackupFilename("articles-20240315-13:13:17Z.json.gz")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})

	t.Run("missing time part", func(t *testing.T) {
		_, err := ParseBackupFilename("articles-20240315Z.json.gz")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})

	t.Run("missing date part", func(t *testing.T) {
		_, err := ParseBackupFilename("articles-131317Z.json.gz")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})

	t.Run("too many digits in date", func(t *testing.T) {
		_, err := ParseBackupFilename("articles-202403151-131317Z.json.gz")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})

	t.Run("too few digits in time", func(t *testing.T) {
		_, err := ParseBackupFilename("articles-20240315-1313Z.json.gz")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})

	t.Run("alpha characters in timestamp", func(t *testing.T) {
		_, err := ParseBackupFilename("articles-abcdefghijklmnop-qrstuvZ.json.gz")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})
}
