# Agent Guidelines for savetoink

This repository contains the code for the savetoink application, composed of:

1. Golang API HTTP backend in [cmd/lambda](cmd/lambda).
2. Frontend web application in [cmd/webapp](cmd/webapp).
3. Browser extension in [cmd/extension](cmd/extension).
4. Landing page website in [cmd/website](cmd/website).

## Development Guidelines

- APIs currently unstable so no need to keep any backward compatibility
- **ALWAYS** run `just lint test` and fix issues before considering a change ready for user review.
- **NEVER** ignore linting errors via `//nolint` statements or similar tricks without prompting the user for permission.
- prefer lowercase log and error messages

## API Documentation

- The OpenAPI specification is maintained in [internal/server/openapi.yaml](internal/server/openapi.yaml)
- **When modifying API routes, handlers, or request/response models in `internal/server/`, you MUST update the OpenAPI specification to reflect the changes**
- The OpenAPI spec should always accurately reflect the current state of the API implementation
