[![Go Report Card](https://goreportcard.com/badge/github.com/shaftoe/savetoink)](https://goreportcard.com/report/github.com/shaftoe/savetoink) [![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=savetoink_savetoink&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=savetoink_savetoink) [![codecov](https://codecov.io/gh/savetoink/savetoink/graph/badge.svg?token=1UHXX9P625)](https://codecov.io/gh/savetoink/savetoink) [![Coverage Status](https://coveralls.io/repos/github/savetoink/savetoink/badge.svg)](https://coveralls.io/github/savetoink/savetoink) [![Api Reference](https://img.shields.io/badge/API-Reference-orange?link=https%3A%2F%2Fwww.saveto.ink%2Fapi-reference%2F)](https://www.saveto.ink/api-reference)

# Save To Ink

Self-hosted read-later service with native e-readers delivery. Save articles in the cloud, send to your e-reader, keep them forever. Open-source alternative to Pocket + Send-to-Kindle.

**DISCLAIMER**: This project is under development and not affiliated with Amazon or Kindle nor Kobo or other e-readers. Use at your own risk.

## Features

- Fetch web pages (articles, blog posts, etc), strip markup with [go-trafilatura](https://github.com/markusmobius/go-trafilatura) and save main readable content as HTML
- Convert content to EPUB format with [go-epub](https://github.com/go-shiori/go-epub) for e-reader devices
- Optionally send to reader devices like Kindle, Kobo, etc. via email backend (only [MailJet](https://www.mailjet.com/) supported at the moment)
- Deploys anywhere with Docker, also as Serverless AWS applicatio. Pluggable storage backend (Sqlite/DynamoDB)
- Run as self-hosted web application (Web Frontend + API server) or as standalone [CLI tool](#cli-tool)
- Browser extension (Chrome, Firefox)to easy save/send pages
- [REST APIs](https://www.saveto.ink/api-reference/) for programmatic access

## CLI Tool

The CLI tool allows you to convert web articles to EPUB format and send them to your reader device directly from the terminal.

### Docker

You can run the CLI tool using Docker:

```bash
alias savetoink="docker run --rm ghcr.io/savetoink/savetoink-cli:latest"
savetoink version
```

Frontend and API server are also available as a Docker image:

```bash
# HTTP server
docker run --rm --env-file .env -p 8080:8080 -v $(pwd)/data:/app/data ghcr.io/savetoink/savetoink-http:latest

# Web frontend
docker run --rm --env-file .env -p 3000:3000 ghcr.io/savetoink/savetoink-webapp:latest
```

**Note**: When using SQLite as storage backend, you must mount a volume at `/app/data` to persist the database between container restarts. The example above mounts a `data` directory from your current working directory.

You can customize the database location using the `SAVETOINK_SQLITE_PATH` environment variable (default: `savetoink.db` in the container's working directory).

### Installation

You can build and install the CLI tool using Go:

```bash
go install github.com/shaftoe/savetoink/cli/savetoink@latest
```

Or build locally using Just:

```bash
just build-cli
```

### Usage

**Send directly to Kindle via email (requires MailJet credentials as environment variables):**

```bash
savetoink send https://example.com --dest-email my-kindle@kindle.com
```

**Convert a URL to EPUB (save locally):**

```bash
savetoink convert https://example.com
```

**Specify an output file:**

```bash
savetoink convert https://example.com -o my-book.epub
```

### Examples

**Save to local file:**

```bash
$ savetoink convert https://golang.org/doc/effective_go.html -o effective_go.epub
✓ EPUB saved to: effective_go.epub
```

**Send to Kindle via email:**

```bash
$ savetoink send https://golang.org/doc/effective_go.html --dest-email my-kindle@kindle.com
✓ Article sent to e-reader device at my-kindle@kindle.com (email ID: fbbaf039-887a-4c86-9a8a-759562592599)"
```

## Development

### Conventions

- this project follows the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) convention (or at least tries to)
- before releasing (pushing to GitHub `master` branch), bump the version using `just bump-version`

### Prerequisites

1. Install [Just command runner](https://just.systems/)
1. Set required environment variables in `.env` (see [backend/lib/config/config.go](backend/lib/config/config.go) for details)
1. Install AWS CLI and configure credentials (optional, only for AWS Lambda deployment)

### Frontend

- SSR SvelteKit app, running as CloudFlare worker or self hosted
- WXT browser extension

To develop the frontend app locally:

```bash
just webapp
```

To develop the extension locally

```bash
just extension
# or
just extension-firefox
```

### Backend

- generic Go HTTP server with AWS DynamoDB backend
- or deployed as AWS Lambda Function (with [HTTP adapter](https://github.com/akrylysov/algnhsa) + CloudFront for custom domain
- pluggable database backend (DynamoDB or SQLite)
- pluggable user backend
  -  single-user shared API key
  -  multi-user with [Auth0](https://auth0.com/)
- pluggable send email backend
  - currently only [MailJet](https://www.mailjet.com/) supported

#### Deployment

To run locally:

```bash
just http
```

To deploy to AWS Lambda:

```bash
# Full AWS Lambda deployment
just deploy

# Destroy AWS Lambda infrastructure
just destroy
```

## License

See [LICENSE](LICENSE.md) file for details.
