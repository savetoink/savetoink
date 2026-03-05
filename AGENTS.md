# Agent Guidelines for savetoink

This repository contains the code for the savetoink application, composed of:

- Golang API HTTP backend in [cmd/http](cmd/http).
  - Lambda function wrapper in [cmd/lambda](cmd/lambda).
- Frontend SvelteKit application in [cmd/web/webapp](cmd/web/webapp).
- Browser WXT extension in [cmd/web/extension](cmd/web/extension).
- Landing page Astro website in [cmd/web/website](cmd/web/website).
- Shared TypeScript web library in [cmd/web/shared](cmd/web/shared).

## Development Guidelines

- APIs currently unstable so no need to keep any backward compatibility
- **ALWAYS** run `just lint test` and fix issues before considering a change ready for user review.
- **NEVER** ignore linting errors via `//nolint` statements or similar tricks without prompting the user for permission.
- prefer lowercase log and error messages

## API Documentation

- The OpenAPI specification is maintained in [internal/server/openapi.yaml](internal/server/openapi.yaml)
- **When modifying API routes, handlers, or request/response models in `internal/server/`, you MUST update the OpenAPI specification to reflect the changes**
- The OpenAPI spec should always accurately reflect the current state of the API implementation
