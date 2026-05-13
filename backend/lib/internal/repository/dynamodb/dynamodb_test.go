package repository

import (
	"context"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	s.repositories = NewDynamoDB(nil, "test-articles", "test-user-profiles", "test-sends", "test-article-tags")
	s.repositories.client = dynamoDBClient
}

func (s *DynamoDBRepositoryTestSuite) TearDownSuite() {
	if err := s.dbContainer.Terminate(s.ctx); err != nil {
		log.Printf("failed to terminate container: %v", err)
	}
}

func (s *DynamoDBRepositoryTestSuite) SetupTest() {
	// Clean up tables before each test to ensure test isolation
	_, _ = s.client.DeleteTable(s.ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String("test-articles"),
	})
	_, _ = s.client.DeleteTable(s.ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String("test-article-tags"),
	})
	_, _ = s.client.DeleteTable(s.ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String("test-user-profiles"),
	})
	_, _ = s.client.DeleteTable(s.ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String("test-sends"),
	})
	err := testhelpers.SetupAllTables(s.ctx, s.client)
	if err != nil {
		log.Printf("failed to setup tables for test: %v", err)
	}
}

func TestNewDynamoDB(t *testing.T) {
	tests := []struct {
		name              string
		articlesTableName string
		profileTableName  string
		sendsTableName    string
		articleTagsTable  string
		shouldPanic       bool
		expectedPanicMsg  string
	}{
		{
			name:              "valid table names",
			articlesTableName: testArticlesTable,
			profileTableName:  testProfilesTable,
			sendsTableName:    testSendsTable,
			articleTagsTable:  testArticleTagsTable,
			shouldPanic:       false,
		},
		{
			name:              "empty articles table name",
			articlesTableName: "",
			profileTableName:  testProfilesTable,
			sendsTableName:    testSendsTable,
			articleTagsTable:  testArticleTagsTable,
			shouldPanic:       true,
			expectedPanicMsg:  testArticlesRequired,
		},
		{
			name:              "empty profile table name",
			articlesTableName: testArticlesTable,
			profileTableName:  "",
			sendsTableName:    testSendsTable,
			articleTagsTable:  testArticleTagsTable,
			shouldPanic:       true,
			expectedPanicMsg:  "user profile table name is required",
		},
		{
			name:              "empty sends table name",
			articlesTableName: testArticlesTable,
			profileTableName:  testProfilesTable,
			sendsTableName:    "",
			articleTagsTable:  testArticleTagsTable,
			shouldPanic:       true,
			expectedPanicMsg:  "sends table name is required",
		},
		{
			name:              "all empty table names",
			articlesTableName: "",
			profileTableName:  "",
			sendsTableName:    "",
			articleTagsTable:  "",
			shouldPanic:       true,
			expectedPanicMsg:  testArticlesRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assert.PanicsWithValue(t, tt.expectedPanicMsg, func() {
					NewDynamoDB(nil, tt.articlesTableName, tt.profileTableName, tt.sendsTableName, tt.articleTagsTable)
				})
			} else {
				db := NewDynamoDB(nil, tt.articlesTableName, tt.profileTableName, tt.sendsTableName, tt.articleTagsTable)
				assert.NotNil(t, db)
				assert.Equal(t, tt.articlesTableName, db.articleTableName)
				assert.Equal(t, tt.profileTableName, db.profileTableName)
				assert.Equal(t, tt.sendsTableName, db.sendsTableName)
				assert.Equal(t, tt.articleTagsTable, db.articleTagsTableName)
			}
		})
	}
}

const (
	dynamoDBTestAccount  = "test-account"
	testArticlesTable    = "articles"
	testProfilesTable    = "profiles"
	testSendsTable       = "sends"
	testArticleTagsTable = "article-tags"
	testArticlesRequired = "articles table name is required"
)

func TestUnmarshalItem(t *testing.T) {
	t.Run("successful unmarshal", func(t *testing.T) {
		item := map[string]types.AttributeValue{
			attributeNameAccount: &types.AttributeValueMemberS{Value: dynamoDBTestAccount},
			attributeNameID:      &types.AttributeValueMemberS{Value: "test-id"},
		}

		var article model.Article
		err := unmarshalItem(item, &article, "article")

		assert.NoError(t, err)
		assert.Equal(t, dynamoDBTestAccount, article.Account)
		assert.Equal(t, "test-id", article.ID)
	})

	t.Run("unmarshal error - nil target", func(t *testing.T) {
		item := map[string]types.AttributeValue{
			attributeNameAccount: &types.AttributeValueMemberS{Value: dynamoDBTestAccount},
		}

		err := unmarshalItem(item, nil, "article")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal article")
	})

	t.Run("unmarshal error - invalid binary data", func(t *testing.T) {
		item := map[string]types.AttributeValue{
			attributeNameAccount: &types.AttributeValueMemberB{Value: []byte{0xff, 0xfe, 0xfd}},
		}

		var article model.Article
		err := unmarshalItem(item, &article, "article")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal article")
	})
}
