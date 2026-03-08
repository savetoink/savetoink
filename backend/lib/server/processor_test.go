package server

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/processor"
	"github.com/shaftoe/savetoink/backend/lib/processor/lambda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProcessor_LambdaConfigured(t *testing.T) {
	cfg := &config.Config{
		ProcessArticleLambda: "process-article-function",
		AWSConfig:            &aws.Config{},
	}
	svc := &mockService{}

	proc := newProcessor(cfg, svc)

	assert.IsType(t, &lambda.Processor{}, proc)
	lambdaProc, ok := proc.(*lambda.Processor)
	require.True(t, ok)
	assert.NotNil(t, lambdaProc)
}

func TestNewProcessor_LambdaNotConfigured(t *testing.T) {
	cfg := &config.Config{
		ProcessArticleLambda: "",
		AWSConfig:            &aws.Config{},
	}
	svc := &mockService{}

	proc := newProcessor(cfg, svc)

	assert.IsType(t, &processor.LocalProcessor{}, proc)
	localProc, ok := proc.(*processor.LocalProcessor)
	require.True(t, ok)
	assert.NotNil(t, localProc)
}

func TestNewProcessor_ImplementsProcessorInterface(t *testing.T) {
	cfg := &config.Config{
		ProcessArticleLambda: "",
		AWSConfig:            &aws.Config{},
	}
	svc := &mockService{}

	proc := newProcessor(cfg, svc)

	var _ = proc
	assert.NotNil(t, proc)
}
