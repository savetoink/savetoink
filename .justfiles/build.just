# Build CLI binary
build-cli:
    CGO_ENABLED=0 go build -o bin/savetoink ./cli/savetoink

build-http:
    CGO_ENABLED=0 go build -o bin/savetoink ./backend/cmd/http

# Build CLI and convert URL into EPUB
run *ARGS: build-cli
    ./bin/savetoink convert {{ ARGS }}

build-cli-docker:
    docker build . -f infra/Dockerfile.cli -t ghcr.io/savetoink/savetoink-cli:local

build-http-docker:
    docker build . -f infra/Dockerfile.http -t ghcr.io/savetoink/savetoink-http:local

push-cli-docker:
    docker tag ghcr.io/savetoink/savetoink-cli:$(cat VERSION) ghcr.io/savetoink/savetoink-cli:latest
    docker push ghcr.io/savetoink/savetoink-cli:latest

push-http-docker:
    docker tag ghcr.io/savetoink/savetoink-http:$(cat VERSION) ghcr.io/savetoink/savetoink-http:latest
    docker push ghcr.io/savetoink/savetoink-http:latest
