# Run linter
[working-directory('backend')]
lint:
    golangci-lint run

test-go:
    go test ./... -short

test-go-coverage:
    go test -coverprofile=coverage.txt ./...

# Run all tests
test: test-go test-webapp test-extension test-website

# Run linter and tests
check: lint test

[working-directory('frontend/webapp')]
test-webapp:
    bun run check
    bun run lint
    bun run test

[working-directory('frontend/extension')]
test-extension:
    bun run check
    bun run lint

[working-directory('frontend/website')]
test-website:
    bun run check
    bun run lint
