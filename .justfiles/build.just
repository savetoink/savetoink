# Build CLI binary
build-cli:
    CGO_ENABLED=0 go build -o bin/savetoink ./cli/savetoink

build-http:
    CGO_ENABLED=0 go build -o bin/savetoink ./backend/cmd/http

# Build CLI and convert URL into EPUB
run *ARGS: build-cli
    ./bin/savetoink convert {{ ARGS }}

build-cli-docker:
    docker build . -f infra/Dockerfile.cli -t ghcr.io/savetoink/savetoink-cli:$(cat VERSION)

build-http-docker:
    docker build . -f infra/Dockerfile.http -t ghcr.io/savetoink/savetoink-http:$(cat VERSION)

push-cli-docker:
    docker tag ghcr.io/savetoink/savetoink-cli:$(cat VERSION) ghcr.io/savetoink/savetoink-cli:latest
    docker push ghcr.io/savetoink/savetoink-cli:latest

push-http-docker:
    docker tag ghcr.io/savetoink/savetoink-http:$(cat VERSION) ghcr.io/savetoink/savetoink-http:latest
    docker push ghcr.io/savetoink/savetoink-http:latest

build-webapp-docker:
    docker build . --build-arg PUBLIC_AUTH_BACKEND --build-arg PUBLIC_API_URL --build-arg PUBLIC_APP_URL --build-arg PUBLIC_IS_DEV_WORKER --build-arg PUBLIC_AUTH0_CLIENT_ID --build-arg PUBLIC_AUTH0_DOMAIN --build-arg PUBLIC_AUTH0_AUDIENCE --build-arg COOKIE_SECRET --build-arg PUBLIC_SENTRY_DSN --build-arg PUBLIC_SENTRY_ENVIRONMENT -f infra/Dockerfile.webapp -t ghcr.io/savetoink/savetoink-webapp:$(cat VERSION)

push-webapp-docker:
    docker tag ghcr.io/savetoink/savetoink-webapp:$(cat VERSION) ghcr.io/savetoink/savetoink-webapp:latest
    docker push ghcr.io/savetoink/savetoink-webapp:latest
