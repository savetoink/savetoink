package consts

// Storage constants.
const (
	// DynamoDBBatchSize is the maximum number of items in a BatchWriteItem operation.
	DynamoDBBatchSize = 25

	// DynamoDBGSIName is the name of the Global Secondary Index for sorting articles by creation date.
	DynamoDBGSIName = "AccountCreatedAtIndex"
)
