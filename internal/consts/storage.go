package consts

// Storage constants.
const (
	// DynamoDBBatchSize is the maximum number of items in a BatchWriteItem operation.
	DynamoDBBatchSize = 25

	// DynamoDBGSIName is the name of the Global Secondary Index for sorting articles by creation date.
	DynamoDBGSIName = "AccountCreatedAtIndex"

	// DynamoDBSendsArticleIDIndex is the name of the Global Secondary Index for querying sends by article ID.
	DynamoDBSendsArticleIDIndex = "ArticleIdIndex"

	// DynamoDBSendsAccountSentAtIndex is the GSI for querying sends by account and sent date.
	DynamoDBSendsAccountSentAtIndex = "AccountSentAtIndex"

	// DynamoDBDeviceEmailIndex is the GSI for querying user profiles by device email.
	DynamoDBDeviceEmailIndex = "DeviceEmailIndex"
)
