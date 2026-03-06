# Build CLI binary
build-cli:
    go build -ldflags "{{ build_flags }}=$(just version)" -o bin/savetoink ./backend/cmd/cli

# Build CLI and convert URL into EPUB
run *ARGS: build-cli
    ./bin/savetoink convert {{ ARGS }}

# Build Lambda binary for Linux
build-lambda:
    GOOS=linux GOARCH=amd64 go build -ldflags "{{ build_flags }}=$(just version)" -o bin/bootstrap ./backend/cmd/lambda

# Build Lambda zip for deployment
[working-directory('bin')]
build-lambda-zip: build-lambda
    zip {{ lambda_archive }} bootstrap
