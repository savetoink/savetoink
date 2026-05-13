package consts

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testDynamoDBGSIName                 = "DynamoDBGSIName"
	testDynamoDBAccountFavoriteIndex    = "DynamoDBAccountFavoriteIndex"
	testDynamoDBSendsArticleIDIndex     = "DynamoDBSendsArticleIDIndex"
	testDynamoDBSendsAccountSentAtIndex = "DynamoDBSendsAccountSentAtIndex"
	testDynamoDBDeviceEmailIndex        = "DynamoDBDeviceEmailIndex"
)

func TestDynamoDB_BatchSize(t *testing.T) {
	assert.Greater(t, DynamoDBBatchSize, 0, "DynamoDBBatchSize should be positive")
	assert.LessOrEqual(t, DynamoDBBatchSize, 25, "DynamoDBBatchSize should not exceed AWS limit")
	assert.Equal(t, 25, DynamoDBBatchSize, "DynamoDBBatchSize should be 25")
}

func TestDynamoDB_GSI_Names(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{testDynamoDBGSIName, DynamoDBGSIName, "AccountCreatedAtIndex"},
		{testDynamoDBAccountFavoriteIndex, DynamoDBAccountFavoriteIndex, "AccountFavoriteIndex"},
		{testDynamoDBSendsArticleIDIndex, DynamoDBSendsArticleIDIndex, "ArticleIdIndex"},
		{testDynamoDBSendsAccountSentAtIndex, DynamoDBSendsAccountSentAtIndex, "AccountSentAtIndex"},
		{testDynamoDBDeviceEmailIndex, DynamoDBDeviceEmailIndex, "DeviceEmailIndex"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.value, "GSI name constant should not be empty")
			assert.Equal(t, tt.expected, tt.value, "GSI name constant should have expected value")
		})
	}
}

func TestDynamoDB_GSI_NamingPattern(t *testing.T) {
	tests := []struct {
		name string
		gsi  string
	}{
		{testDynamoDBGSIName, DynamoDBGSIName},
		{testDynamoDBAccountFavoriteIndex, DynamoDBAccountFavoriteIndex},
		{testDynamoDBSendsArticleIDIndex, DynamoDBSendsArticleIDIndex},
		{testDynamoDBSendsAccountSentAtIndex, DynamoDBSendsAccountSentAtIndex},
		{testDynamoDBDeviceEmailIndex, DynamoDBDeviceEmailIndex},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, strings.HasSuffix(tt.gsi, "Index"), "GSI name should end with 'Index'")
		})
	}
}

func TestStorageBackend_Constants(t *testing.T) {
	tests := []struct {
		name     string
		constant StorageBackend
		value    string
	}{
		{"DynamoDB", StorageBackendDynamoDB, string(StorageBackendDynamoDB)},
		{"SQLite", StorageBackendSQLite, string(StorageBackendSQLite)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.constant, "Storage backend constant should not be empty")
			assert.Equal(t, tt.value, string(tt.constant), "Storage backend constant should have expected value")
		})
	}
}

func TestStorageBackend_String_Conversion(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want StorageBackend
	}{
		{string(StorageBackendDynamoDB), string(StorageBackendDynamoDB), StorageBackendDynamoDB},
		{string(StorageBackendSQLite), string(StorageBackendSQLite), StorageBackendSQLite},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StorageBackend(tt.env)
			assert.Equal(t, tt.want, got)
		})
	}
}
