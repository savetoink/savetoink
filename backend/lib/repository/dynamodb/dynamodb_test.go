package repository

import (
	"context"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	testhelpers "github.com/shaftoe/savetoink/backend/lib/repository/testhelpers"
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
