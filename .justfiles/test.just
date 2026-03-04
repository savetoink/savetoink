# Run linter
lint:
    golangci-lint run

# Run tests (skip DynamoDB integration tests)
test-go:
    go test ./... -short

# Run all tests
test: test-go test-webapp

# Run linter and tests
check: lint test

[working-directory('cmd/web/webapp')]
test-webapp:
    npm run check
    npm run lint
    npm run test
