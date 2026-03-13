package consts

import "time"

// StorageBackend defines the storage backend type.
type StorageBackend string

const (
	// StorageBackendDynamoDB indicates AWS DynamoDB storage backend.
	StorageBackendDynamoDB StorageBackend = "dynamodb"
	// StorageBackendSQLite indicates SQLite storage backend.
	StorageBackendSQLite StorageBackend = "sqlite"
)

// Storage constants.
const (
	// DynamoDBBatchSize is maximum number of items in a BatchWriteItem operation.
	DynamoDBBatchSize = 25

	// DynamoDBGSIName is name of the Global Secondary Index for sorting articles by creation date.
	DynamoDBGSIName = "AccountCreatedAtIndex"

	// DynamoDBSendsArticleIDIndex is name of the Global Secondary Index for querying sends by article ID.
	DynamoDBSendsArticleIDIndex = "ArticleIdIndex"

	// DynamoDBSendsAccountSentAtIndex is the GSI for querying sends by account and sent date.
	DynamoDBSendsAccountSentAtIndex = "AccountSentAtIndex"

	// DynamoDBDeviceEmailIndex is the GSI for querying user profiles by device email.
	DynamoDBDeviceEmailIndex = "DeviceEmailIndex"

	// SqliteInitTimeout is the timeout for initializing the SQLite database.
	SqliteInitTimeout = time.Second
)
