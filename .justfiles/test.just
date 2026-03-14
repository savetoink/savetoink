# Run linters
lint: lint-cfn lint-go

lint-cfn:
    cfn-lint -t infra/*.yaml

lint-go:
    CGO_ENABLED=0 golangci-lint run backend/... cli/...

test-go: test-go-backend test-go-cli

test-go-backend:
    CGO_ENABLED=0 go test ./backend/...

test-go-cli:
    CGO_ENABLED=0 go test ./cli/...

test-go-coverage:
    CGO_ENABLED=0 go test -coverprofile=coverage.txt ./backend/...

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
