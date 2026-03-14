package repository

import (
	"context"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	testhelpers "github.com/shaftoe/savetoink/backend/lib/internal/repository/testhelpers"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	tcdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
)

type DynamoDBRepositoryTestSuite struct {
	suite.Suite
	ctx          context.Context
	dbContainer  *tcdynamodb.DynamoDBContainer
	client       *dynamodb.Client
	repositories *DynamoDB
}

func TestDynamoDBRepositoryTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	suite.Run(t, new(DynamoDBRepositoryTestSuite))
}

func (s *DynamoDBRepositoryTestSuite) SetupSuite() {
	s.ctx = context.Background()

	var err error
	s.dbContainer, err = testhelpers.CreateDynamoDBContainer(s.ctx)
	if err != nil {
		log.Fatalf("failed to create dynamodb container: %v", err)
	}

	dynamoDBClient, err := testhelpers.CreateDynamoDBClient(s.ctx, s.dbContainer)
	if err != nil {
		log.Fatalf("failed to create dynamodb client: %v", err)
	}

	err = testhelpers.SetupAllTables(s.ctx, dynamoDBClient)
	if err != nil {
		log.Fatalf("failed to setup tables: %v", err)
	}

	s.client = dynamoDBClient
	s.repositories = NewDynamoDB(nil, "test-articles", "test-user-profiles", "test-sends")
	s.repositories.client = dynamoDBClient
}

func (s *DynamoDBRepositoryTestSuite) TearDownSuite() {
	if err := s.dbContainer.Terminate(s.ctx); err != nil {
		log.Printf("failed to terminate container: %v", err)
	}
}

func TestNewDynamoDB(t *testing.T) {
	tests := []struct {
		name              string
		articlesTableName string
		profileTableName  string
		sendsTableName    string
		shouldPanic       bool
		expectedPanicMsg  string
	}{
		{
			name:              "valid table names",
			articlesTableName: "articles",
			profileTableName:  "profiles",
			sendsTableName:    "sends",
			shouldPanic:       false,
		},
		{
			name:              "empty articles table name",
			articlesTableName: "",
			profileTableName:  "profiles",
			sendsTableName:    "sends",
			shouldPanic:       true,
			expectedPanicMsg:  "articles table name is required",
		},
		{
			name:              "empty profile table name",
			articlesTableName: "articles",
			profileTableName:  "",
			sendsTableName:    "sends",
			shouldPanic:       true,
			expectedPanicMsg:  "user profile table name is required",
		},
		{
			name:              "empty sends table name",
			articlesTableName: "articles",
			profileTableName:  "profiles",
			sendsTableName:    "",
			shouldPanic:       true,
			expectedPanicMsg:  "sends table name is required",
		},
		{
			name:              "all empty table names",
			articlesTableName: "",
			profileTableName:  "",
			sendsTableName:    "",
			shouldPanic:       true,
			expectedPanicMsg:  "articles table name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assert.PanicsWithValue(t, tt.expectedPanicMsg, func() {
					NewDynamoDB(nil, tt.articlesTableName, tt.profileTableName, tt.sendsTableName)
				})
			} else {
				db := NewDynamoDB(nil, tt.articlesTableName, tt.profileTableName, tt.sendsTableName)
				assert.NotNil(t, db)
				assert.Equal(t, tt.articlesTableName, db.articleTableName)
				assert.Equal(t, tt.profileTableName, db.profileTableName)
				assert.Equal(t, tt.sendsTableName, db.sendsTableName)
			}
		})
	}
}

func TestUnmarshalItem(t *testing.T) {
	t.Run("successful unmarshal", func(t *testing.T) {
		item := map[string]types.AttributeValue{
			"account": &types.AttributeValueMemberS{Value: "test-account"},
			"id":      &types.AttributeValueMemberS{Value: "test-id"},
		}

		var article model.Article
		err := unmarshalItem(item, &article, "article")

		assert.NoError(t, err)
		assert.Equal(t, "test-account", article.Account)
		assert.Equal(t, "test-id", article.ID)
	})

	t.Run("unmarshal error - nil target", func(t *testing.T) {
		item := map[string]types.AttributeValue{
			"account": &types.AttributeValueMemberS{Value: "test-account"},
		}

		err := unmarshalItem(item, nil, "article")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal article")
	})

	t.Run("unmarshal error - invalid binary data", func(t *testing.T) {
		item := map[string]types.AttributeValue{
			"account": &types.AttributeValueMemberB{Value: []byte{0xff, 0xfe, 0xfd}},
		}

		var article model.Article
		err := unmarshalItem(item, &article, "article")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal article")
	})
}
