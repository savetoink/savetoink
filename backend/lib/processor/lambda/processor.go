// Package lambda provides Lambda-based article processing.
package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/processor"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
)

type lambdaInvoker interface {
	Invoke(ctx context.Context, params *lambda.InvokeInput, optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error)
}

type awsLambdaClient struct {
	client *lambda.Client
}

func (a *awsLambdaClient) Invoke(
	ctx context.Context,
	params *lambda.InvokeInput,
	optFns ...func(*lambda.Options),
) (*lambda.InvokeOutput, error) {
	result, err := a.client.Invoke(ctx, params, optFns...)
	if err != nil {
		return nil, fmt.Errorf("lambda invoke failed: %w", err)
	}
	return result, nil
}

// Processor invokes a remote Lambda function for article processing.
type Processor struct {
	functionName  string
	lambdaInvoker lambdaInvoker
}

// NewProcessor creates a new Processor.
func NewProcessor(functionName string, awsCfg *aws.Config) *Processor {
	client := lambda.NewFromConfig(*awsCfg)
	return &Processor{
		functionName:  functionName,
		lambdaInvoker: &awsLambdaClient{client: client},
	}
}

// newProcessorWithInvoker creates a new Processor with a custom invoker (for testing).
func newProcessorWithInvoker(
	functionName string, invoker lambdaInvoker) *Processor { //nolint:unparam // Used for testing only
	return &Processor{
		functionName:  functionName,
		lambdaInvoker: invoker,
	}
}

// StartProcessing invokes a remote Lambda function to process an article.
func (p *Processor) StartProcessing(ctx context.Context, event *content.ProcessArticleEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		logging.AddRequestError(ctx, fmt.Errorf("failed to marshal event payload: %w", err))
		return
	}

	_, err = p.lambdaInvoker.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(p.functionName),
		InvocationType: types.InvocationTypeEvent,
		Payload:        payload,
	})
	if err != nil {
		logging.AddRequestError(ctx, fmt.Errorf("failed to invoke process article lambda: %w", err))
	}

	logging.AddLogAttr(ctx, slog.String("processed_by", p.functionName))
}

var _ processor.Processor = (*Processor)(nil)
