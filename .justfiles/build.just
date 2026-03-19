# Build CLI binary
build-cli:
    CGO_ENABLED=0 go build -o bin/savetoink ./cli/savetoink

# Build HTTP binary
build-http:
    CGO_ENABLED=0 go build -o bin/savetoink ./backend/cmd/http

build-docker APP:
    docker build . --build-arg APP={{ APP }} -f backend/Dockerfile -t ghcr.io/savetoink/savetoink-{{ APP }}:local

# Build CLI Docker image
build-cli-docker:
    just build-docker cli

# Build HTTP Docker image
build-http-docker:
    just build-docker http

build-webapp-docker:
    docker build . \
    --build-arg PUBLIC_AUTH_BACKEND \
    --build-arg PUBLIC_IS_DEV_WORKER \
    -f frontend/Dockerfile \
    -t ghcr.io/savetoink/savetoink-webapp:local

[working-directory('frontend/webapp')]
build-webapp:
    @bun run build
