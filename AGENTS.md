# Agent Guidelines for savetoink

This repository contains the code for the savetoink application, composed of:

- Golang API HTTP backend in [backend/cmd/http](backend/cmd/http).
  - Lambda function wrapper in [backend/cmd/lambda](backend/cmd/lambda).
- Golang CLI tool in [cli/savetoink](cli/savetoink).
- Frontend SvelteKit application in [frontend/webapp](frontend/webapp).
- Browser WXT extension in [frontend/extension](frontend/extension).
- Landing page Astro website in [frontend/website](frontend/website).
- Shared TypeScript web library in [frontend/shared](frontend/shared).

## Development Guidelines

- APIs currently unstable so no need to keep any backward compatibility
- **ALWAYS** add new (unit) tests for new features.
- **ALWAYS** run `just check` and fix issues before considering a change ready for user review. Strive for 100% test coverage, use `go tool cover` to check coverage.
- **NEVER** ignore linting errors via `//nolint` statements or similar tricks without prompting the user for permission.
- prefer lowercase log and error messages

### Backend / SSR

- prefer wide event logging: collect all relevant fields in a single event rather than multiple events, log once when all the processing is done
- prefer keeping constant values in the dedicated `consts` package

## API Documentation

- The OpenAPI specification is maintained in [backend/lib/server/openapi.yaml](backend/lib/server/openapi.yaml)
- **When modifying API routes, handlers, or request/response models in `backend/lib/server/`, you MUST update the OpenAPI specification to reflect the changes**
- The OpenAPI spec should always accurately reflect the current state of the API implementation

## DynamoDB Integration Tests

- The DynamoDB Local schema for integration tests is defined in [backend/lib/repository/testhelpers/setup.go](backend/lib/repository/testhelpers/setup.go)
- **When modifying DynamoDB table schemas in [infra/api.yaml](infra/api.yaml), you MUST keep the test schema in sync with the production schema**
- The testcontainers integration tests use a local DynamoDB instance; ensure table definitions, attribute types, key schemas, and global secondary indexes match the production CloudFormation template
