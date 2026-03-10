# Run linters
lint:
    golangci-lint run cli/... backend/...

lint-go:
    golangci-lint run backend/... cli/...

test-go: test-go-backend test-go-cli

test-go-backend:
    go test ./backend/...

test-go-cli:
    go test ./cli/...

test-go-coverage:
    go test -coverprofile=coverage.txt ./backend/...

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
