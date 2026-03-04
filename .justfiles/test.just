# Run linter
lint:
    golangci-lint run

# Run tests (skip DynamoDB integration tests)
test-go:
    go test ./... -short

# Run all tests
test: test-go test-webapp test-extension test-website

# Run linter and tests
check: lint test

[working-directory('cmd/web/webapp')]
test-webapp:
    bun run check
    bun run lint
    bun run test

[working-directory('cmd/web/extension')]
test-extension:
    bun run check
    bun run lint

[working-directory('cmd/web/website')]
test-website:
    bun run check
    bun run lint
