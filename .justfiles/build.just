# Build CLI binary
build-cli:
    go build -o bin/savetoink ./cli/savetoink

# Build CLI and convert URL into EPUB
run *ARGS: build-cli
    ./bin/savetoink convert {{ ARGS }}
