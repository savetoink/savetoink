# Agent Guidelines for savetoink

This repository contains the code for the savetoink application, composed of:

- Golang API HTTP backend in [backend/cmd/http](backend/cmd/http).
  - Lambda function wrapper in [backend/cmd/lambda](backend/cmd/lambda).
- Frontend SvelteKit application in [frontend/webapp](frontend/webapp).
- Browser WXT extension in [frontend/extension](frontend/extension).
- Landing page Astro website in [frontend/website](frontend/website).
- Shared TypeScript web library in [frontend/shared](frontend/shared).

## Development Guidelines

- APIs currently unstable so no need to keep any backward compatibility
- **ALWAYS** run `just lint test` and fix issues before considering a change ready for user review.
- **NEVER** ignore linting errors via `//nolint` statements or similar tricks without prompting the user for permission.
- prefer lowercase log and error messages

## API Documentation

- The OpenAPI specification is maintained in [backend/internal/server/openapi.yaml](backend/internal/server/openapi.yaml)
- **When modifying API routes, handlers, or request/response models in `backend/internal/server/`, you MUST update the OpenAPI specification to reflect the changes**
- The OpenAPI spec should always accurately reflect the current state of the API implementation
