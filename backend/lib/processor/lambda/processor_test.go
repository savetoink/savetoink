package lambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
)

const testFunctionName = "test-function"

func TestNewProcessor(t *testing.T) {
	awsCfg := aws.Config{
		Region: "us-east-1",
	}
	processor := NewProcessor(testFunctionName, &awsCfg)

	assert.NotNil(t, processor)
	assert.Equal(t, testFunctionName, processor.functionName)
	assert.NotNil(t, processor.lambdaClient)
}
