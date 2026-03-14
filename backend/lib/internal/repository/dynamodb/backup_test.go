package repository

import (
	"bytes"
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
		s3Client:     mockS3,
		bucket:       "test-bucket",
		logger:       logger,
	}

	result := br.BackupTable(context.Background(), "test-table")

	require.NoError(t, result.Error)
	assert.Contains(t, result.Key, "backups/test-table-")
	assert.Contains(t, result.Key, ".json")
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
		s3Client:     mockS3,
		bucket:       "test-bucket",
		logger:       logger,
	}

	result := br.BackupTable(context.Background(), "test-table")

	require.NoError(t, result.Error)
	require.Contains(t, result.Key, "backups/test-table-")
	require.Contains(t, result.Key, ".json")

	parsedTime, err := time.Parse("20060102-150405", result.Key[len("backups/test-table-"):len(result.Key)-len(".json")])
	require.NoError(t, err)
	require.NotNil(t, parsedTime)
}

const (
	testBackupJSON = `{
		"timestamp": "2024-03-15T13:13:17Z",
		"table_name": "test-articles",
		"item_count": 2,
		"items": [
			{"account": {"S": "account1"}, "id": {"S": "id1"}},
			{"account": {"S": "account1"}, "id": {"S": "id2"}}
		]
	}`

	testSingleItemBackupJSON = `{
		"timestamp": "2024-03-15T13:13:17Z",
		"table_name": "test-articles",
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
				return &s3.GetObjectOutput{
					Body: io.NopCloser(bytes.NewReader([]byte(backupJSON))),
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

		result := br.RestoreTable(context.Background(), "articles-20240315-131317.json", true)

		require.NoError(t, result.Error)
		assert.Equal(t, "articles", result.TableName)
		assert.Equal(t, "articles-20240315-131317.json", result.BackupName)
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
				return &s3.GetObjectOutput{
					Body: io.NopCloser(bytes.NewReader([]byte(backupJSON))),
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

		result := br.RestoreTable(context.Background(), "articles-20240315-131317.json", false)

		require.NoError(t, result.Error)
		assert.Equal(t, "articles", result.TableName)
		assert.Equal(t, "articles-20240315-131317.json", result.BackupName)
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
		assert.Contains(t, result.Error.Error(), "backup name must end with .json")
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

		result := br.RestoreTable(context.Background(), "articles-20240315-131317.json", false)

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
				return &s3.GetObjectOutput{
					Body: io.NopCloser(bytes.NewReader([]byte(backupJSON))),
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

		result := br.RestoreTable(context.Background(), "articles-20240315-131317.json", false)

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
				return &s3.GetObjectOutput{
					Body: io.NopCloser(bytes.NewReader([]byte(backupJSON))),
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

		result := br.RestoreTable(context.Background(), "articles-20240315-131317.json", false)

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

		tableName, err := br.extractTableName("articles-20240315-131317.json")
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

		tableName, err := br.extractTableName("savetoink-articles-20240315-131317.json")
		require.NoError(t, err)
		assert.Equal(t, "articles", tableName)
	})

	t.Run("missing .json extension", func(t *testing.T) {
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
		assert.Contains(t, err.Error(), "backup name must end with .json")
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

		_, err := br.extractTableName("invalid.json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "backup name format should be")
	})
}
