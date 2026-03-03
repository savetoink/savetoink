# Save To Ink

Self-hosted read-later service with native Kindle delivery. Save articles in the cloud, send to your e-reader, keep them forever. Open-source alternative to Pocket + Send-to-Kindle.

**DISCLAIMER**: This project is under development (alpha)and not affiliated with Amazon or Kindle. Use at your own risk.

## Features

- Fetch web articles, strip markup with [go-trafilatura](https://github.com/markusmobius/go-trafilatura) and save main readable content as HTML
- Run as web service (API) or as [CLI tool](#cli-tool)
- Convert content to EPUB format with [go-epub](https://github.com/go-shiori/go-epub) for e-reader devices
- Optionally send directly to Kindle via email backend (only [MailJet](https://www.mailjet.com/) supported at the moment)

### Backend

- generic Go HTTP server
- deployed as AWS Lambda Function (with [HTTP adapter](https://github.com/akrylysov/algnhsa) + CloudFront for custom domain, DynamoDB for storage
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

```bash
# Full deployment
just deploy

# Destroy infrastructure
just destroy
```

## CLI Tool

The CLI tool allows you to convert web articles to EPUB format and send them to your Kindle device directly from the terminal.

### Installation

```bash
go build -o bin/savetoink cmd/cli
```

### Usage

**Convert a URL to EPUB (save locally):**

```bash
./bin/savetoink convert https://example.com
```

**Send directly to Kindle via email:**

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

### Browser Extension

**Install the extension:**

**Temporary Installation (Development):**

1. Open Firefox and navigate to `about:debugging#/runtime/this-firefox`
2. Click "Load Temporary Add-on"
3. Select the `extension/manifest.json` file
4. Configure settings by clicking extension icon → "Configure Settings"
5. Enter your API URL and API Key

**Using the extension:**
1. Click the extension icon on any web page to send it to Kindle
2. Or right-click on any link → "Send to Kindle"
3. Configure API settings via extension icon → "Configure Settings"
4. The extension stores your API key and URL securely using Chrome storage API
5. CORS headers are properly configured for cross-origin requests to your Lambda function

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
