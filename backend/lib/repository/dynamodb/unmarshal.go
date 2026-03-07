package repository

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// unmarshalItem unmarshals a DynamoDB item with consistent error handling.
func unmarshalItem(item map[string]types.AttributeValue, target any, typeName string) error {
	if err := attributevalue.UnmarshalMap(item, target); err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", typeName, err)
	}
	return nil
}
