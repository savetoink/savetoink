# SaveToInk Browser Extension

A browser extension that allows you to save articles to your Kindle with a single click or right-click.

## Features

- **Auth0 Authentication**: Secure login with Auth0 OAuth 2.0 PKCE flow
- **One-click saving**: Send the current page URL to SaveToInk API
- **Context menu integration**: Right-click on any page, link, or selected text to save
- **Cross-browser support**: Works on Chrome, Firefox, and Safari (via WXT framework)
- **Automatic token refresh**: Tokens are automatically refreshed when needed

## Development

### Prerequisites

- Node.js 18+
- Bun or npm
- Auth0 credentials configured in `.env`

### Setup

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your Auth0 credentials:
   ```
   VITE_AUTH0_DOMAIN=your-auth0-domain.auth0.com
   VITE_AUTH0_CLIENT_ID=your-client-id
   VITE_AUTH0_AUDIENCE=your-api-audience
   VITE_API_BASE_URL=https://api.saveto.ink/v1
   ```

### Development

Run in development mode (auto-reload):
```bash
npm run dev
```

For Firefox:
```bash
npm run dev:firefox
```

### Build

Build for production:
```bash
npm run build
```

Build for Firefox:
```bash
npm run build:firefox
```

### Create Distribution Package

```bash
npm run zip
```

## Usage

### Authentication

1. Click the SaveToInk extension icon
2. Click "Login with Auth0"
3. Complete the Auth0 authentication flow
4. You'll be redirected back and logged in

### Saving Articles

**Method 1: Popup**
1. Click the SaveToInk extension icon
2. Click "Send this page"

**Method 2: Context Menu**
1. Right-click anywhere on the page
2. Select "Send to SaveToInk"

**Method 3: Link/Selection**
1. Right-click on a link or selected text
2. Select "Send to SaveToInk"

## Architecture

### Components

- **Background Service Worker**: Handles Auth0 flow, API calls, and context menu
- **Popup UI**: Svelte-based interface for authentication and quick actions
- **Content Script**: Minimal script for page interaction
- **Redirect Handler**: Processes OAuth callback from Auth0

### Key Files

- `src/lib/auth.ts`: Auth0 SDK wrapper and token management
- `src/lib/api.ts`: API client for SaveToInk backend
- `src/entrypoints/background.ts`: Service worker with message handlers
- `src/entrypoints/popup/App.svelte`: Main popup UI

### Storage

Uses `chrome.storage.local` for:
- Auth tokens (access, refresh tokens, expiration)
- User profile information
- Authentication state

## License

See project root LICENSE file.
