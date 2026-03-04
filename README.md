# Save To Ink

Self-hosted read-later service with native Kindle delivery. Save articles in the cloud, send to your e-reader, keep them forever. Open-source alternative to Pocket + Send-to-Kindle.

**DISCLAIMER**: This project is under development (alpha)and not affiliated with Amazon or Kindle. Use at your own risk.

## Features

- Fetch web pages (articles, blog posts, etc), strip markup with [go-trafilatura](https://github.com/markusmobius/go-trafilatura) and save main readable content as HTML
- Run as self-hosted web application (Frontend + API) or as [CLI tool](#cli-tool)
- Convert content to EPUB format with [go-epub](https://github.com/go-shiori/go-epub) for e-reader devices
- Optionally send to reader devices like Kindle, Kobo, etc. via email backend (only [MailJet](https://www.mailjet.com/) supported at the moment)
- Browser extension to easy save/send pages

## Frontend

- SSR SvelteKit app, running as CloudFlare worker or self hosted

## Backend

- generic Go HTTP server with AWS DynamoDB backend
- or deployed as AWS Lambda Function (with [HTTP adapter](https://github.com/akrylysov/algnhsa) + CloudFront for custom domain
- pluggable user backend
  -  single-user shared API key
  -  multi-user with [Auth0](https://auth0.com/)
- pluggable send email backend
  - currently only [MailJet](https://www.mailjet.com/) supported

### Prerequisites

1. Install AWS CLI and configure credentials
1. Install [Just command runner](https://just.systems/)
1. Set required environment variables in `.env` (see [internal/config/config.go](internal/config/config.go) for details)

### Deployment

To run locally:

```bash
just server-http
```

To deploy to AWS Lambda:

```bash
# Full AWS Lambda deployment
just deploy

# Destroy AWS Lambda infrastructure
just destroy
```

## CLI Tool

The CLI tool allows you to convert web articles to EPUB format and send them to your reader device directly from the terminal.

### Installation

```bash
go build -o bin/savetoink cmd/cli
```

### Usage

**Convert a URL to EPUB (save locally):**

```bash
./bin/savetoink convert https://example.com
```

**Send directly to Kindle via email (requires MailJet credentials as environment variables):**

```bash
./bin/savetoink convert https://example.com --send
```

**Specify an output file:**

```bash
./bin/savetoink convert https://example.com -o my-book.epub
```

**Set a custom timeout:**

```bash
./bin/savetoink convert https://example.com -t 1m
```

**Show extracted HTML content (verbose mode):**

```bash
./bin/savetoink convert https://example.com -v
```

### Examples

**Save to local file:**

```bash
$ ./bin/savetoink convert https://golang.org/doc/effective_go.html -o effective_go.epub
Fetching article from: https://golang.org/doc/effective_go.html
Extracted in 828ms
Title: Effective Go
Generating EPUB: effective_go.epub
Generated in 7ms

✓ EPUB saved to: /Users/alex/git/savetoink/effective_go.epub
```

**Send to Kindle via email:**

```bash
$ ./bin/savetoink convert https://golang.org/doc/effective_go.html --send
Fetching article from: https://golang.org/doc/effective_go.html
Extracted in 828ms
Title: Effective Go
Generating EPUB for email...
Generated in 7ms
Sending to Kindle: sender@example.com -> your-kindle@kindle.com
Sent in 245ms
Email sent successfully. Message ID: 1234567890, UUID: abc123-def456-ghi789

✓ Article sent to Kindle
```

## License

See [LICENSE](LICENSE.md) file for details.
