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
- wrap errors with `fmt.Errorf` when passing them to functions that expect an error, prefer custom error types over `errors.New` when possible
- HTTP handlers should use, when available, utils helpers to return errors and such
- **Use the service layer as the single entry point**: All external consumers (CLI, HTTP server, Lambda) should interact with backend functionality through the `service` package

### Package Organization

**Public API** (can be imported from anywhere):
- `backend/lib/config` - Configuration management
- `backend/lib/consts` - Constants
- `backend/lib/logging` - Logging setup
- `backend/lib/model` - Shared data models
- `backend/lib/service` - Main facade (single entry point)
- `backend/lib/server` - HTTP server setup
- `backend/lib/processor` - Article processing orchestration
- `backend/lib/scheduler` - Background task scheduling

**Internal** (only for use within `backend/lib/`):
- `backend/lib/internal/` - All internal packages
  - `backend/lib/internal/content` - Article content extraction
  - `backend/lib/internal/epub` - EPUB generation
  - `backend/lib/internal/service/` - Service implementation details
    - `backend/lib/internal/service/articles` - Article CRUD
    - `backend/lib/internal/service/profile` - User profile management
    - `backend/lib/internal/service/servicetypes` - Service result types
  - `backend/lib/internal/server/` - Server implementation details
    - `backend/lib/internal/server/auth` - HTTP authentication middleware
    - `backend/lib/internal/server/handlers` - HTTP request handlers
    - `backend/lib/internal/server/types` - HTTP request/response types
    - `backend/lib/internal/server/utils` - HTTP utility functions
  - `backend/lib/internal/email/` - Email sending
  - `backend/lib/internal/repository/` - Database operations
  - `backend/lib/internal/auth/` - Authentication context helpers
  - `backend/lib/internal/validation` - Input validation
  - `backend/lib/internal/task/` - Background task runner
  - `backend/lib/internal/apperrors` - Application error types

**Important**: Packages in `backend/lib/internal/` cannot be imported from outside `backend/lib/` (Go's internal directory rule). The service layer provides all necessary functionality.

## API Documentation

- The OpenAPI specification is maintained in [backend/lib/server/openapi.yaml](backend/lib/server/openapi.yaml)
- **When modifying API routes, handlers, or request/response models in `backend/lib/server/`, you MUST update the OpenAPI specification to reflect the changes**
- The OpenAPI spec should always accurately reflect the current state of the API implementation

## DynamoDB Integration Tests

- The DynamoDB Local schema for integration tests is defined in [backend/lib/repository/testhelpers/setup.go](backend/lib/repository/testhelpers/setup.go)
- **When modifying DynamoDB table schemas in [infra/api.yaml](infra/api.yaml), you MUST keep the test schema in sync with the production schema**
- The testcontainers integration tests use a local DynamoDB instance; ensure table definitions, attribute types, key schemas, and global secondary indexes match the production CloudFormation template
