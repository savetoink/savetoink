# PASETO Migration Plan

## Executive Summary

This document outlines the plan to replace JWT tokens with PASETO (Platform-Agnostic SEcurity TOkens) for the Auth0 authentication backend. PASETO is **not a separate auth backend** — Auth0 remains the identity provider. PASETO replaces the token format used between backend and frontend, eliminating direct exposure of Auth0 JWTs to the client.

**No backward compatibility.** The migration is a clean cut — JWT passthrough and validation are removed entirely.

## Current Authentication Architecture

### Current Flow
1. Frontend redirects users to Auth0 for authentication
2. Auth0 returns an authorization code
3. Frontend exchanges the authorization code with backend at `/v1/auth/token`
4. Backend exchanges authorization code with Auth0 for tokens (access_token, id_token, refresh_token)
5. Backend extracts email from ID token and stores it
6. Frontend uses Auth0 access_token (JWT) to authenticate subsequent requests
7. Backend validates JWT tokens using `auth0Middleware` with JWKS validation

### Key Components
- **Backend**: `backend/lib/internal/server/auth/auth.go` - Authentication middleware
- **Backend**: `backend/lib/internal/server/handlers/auth.go` - Token exchange endpoint
- **Backend**: `backend/lib/internal/auth/context.go` - Auth context helpers
- **Backend**: `backend/lib/config/config.go` - Configuration
- **Backend**: `backend/lib/consts/app.go` - Auth backend constants

### Current JWT Dependencies
- `github.com/auth0/go-jwt-middleware/v3` - JWT validation library (to be removed)

## Why PASETO?

### JWT Security Issues
- Algorithm confusion attacks
- Key confusion problems
- Lack of strict implementation guidelines
- "None" algorithm vulnerability
- Claim injection attacks

### PASETO Advantages
- Designed to avoid JWT's security pitfalls
- Simpler, more secure by design
- Versioned protocol (v4 recommended)
- Symmetric encryption (v4.local) keeps claims confidential
- No algorithm confusion
- Standardized security practices

## Architecture: PASETO as Token Layer

Auth0 is the identity provider. PASETO replaces the token format between backend and frontend.

```
AuthBackendAuth0 requires:
  - Auth0 config (domain, audience, client ID, client secret) — identity provider
  - PASETO config (symmetric key, key version) — token layer (REQUIRED)
```

### New Authentication Flow

```
Frontend → Auth0 → Frontend (auth code)
Frontend → Backend /v1/auth/token (exchange code)
Backend → Auth0 (exchange for JWT)        ← Auth0 = identity provider
Auth0 → Backend (JWT tokens)
Backend extracts claims from JWT
Backend generates PASETO tokens from claims
Backend → Frontend (PASETO tokens)        ← Frontend never sees Auth0 JWT
Frontend → Backend (PASETO token in Authorization header)
Backend validates PASETO token using local key  ← No JWKS fetch needed
```

### No Backward Compatibility
- PASETO is **required** for `AuthBackendAuth0` — not optional
- No JWT passthrough, no fallback, no prefix detection
- `go-jwt-middleware` will be removed
- The middleware only validates PASETO tokens

## Migration Strategy

### Phase 1: Foundation ✅ COMPLETE

1. **go-paseto library** ✅ — `aidanwoods.dev/go-paseto` v1.6.0
2. **PASETO package** ✅ — `backend/lib/internal/paseto/` with 95% test coverage
   - `KeyStore` — key management with rotation support
   - `GenerateTokenPair()` — access + refresh token generation
   - `ValidateToken()` — token validation with previous-key fallback
   - `IsPASETOToken()` — prefix detection

### Phase 2: Configuration ✅ COMPLETE

1. **PASETO config fields** ✅ — required for Auth0 backend
   - `SAVETOINK_PASETO_KEY` (base64-encoded 32 bytes)
   - `SAVETOINK_PASETO_KEY_VERSION`
   - Both required when `SAVETOINK_AUTH_BACKEND=auth0`
2. **Validation** ✅ — missing PASETO fields reported alongside missing Auth0 fields

### Phase 3: Implementation ✅ COMPLETE

1. **Update `/v1/auth/token` endpoint** ✅
   - [x] After Auth0 exchange, extract claims (sub, email) from Auth0 response
   - [x] Generate PASETO token pair from claims
   - [x] Return PASETO tokens to frontend (never return Auth0 JWTs)

2. **Replace authentication middleware** ✅
   - [x] Remove `auth0Middleware` (JWT/JWKS validation)
   - [x] Add `pasetoMiddleware` using the PASETO KeyStore
   - [x] Extract account ID from PASETO token claims

3. **Wire up PASETO KeyStore** ✅
   - [x] Create KeyStore during server initialization from config
   - [x] Pass KeyStore to token exchange handler
   - [x] Pass KeyStore to auth middleware

4. **Remove JWT dependency** ✅
   - [x] Remove `go-jwt-middleware` import and usage
   - [x] Remove JWKS provider setup
   - [x] Clean up `auth.go`

5. **Testing** ✅
   - [x] Update handler tests for PASETO token generation
   - [x] Update middleware tests for PASETO validation
   - [x] Update router tests
   - [x] Integration test: full PASETO round-trip (exchange → authenticate)

### Phase 4: Frontend Integration

1. **Token handling** — PASETO tokens use the same `Bearer` pattern, should be transparent
2. **Token refresh** — update if response structure changes
3. **E2E tests**

### Phase 5: Deployment

1. Generate 32-byte PASETO symmetric key
2. Deploy with `SAVETOINK_PASETO_KEY` and `SAVETOINK_PASETO_KEY_VERSION`
3. Monitor authentication success/failure

---

## Technical Implementation Details

### PASETO Package (`backend/lib/internal/paseto/`)

```go
type TokenClaims struct {
    Subject    string  // Account ID from Auth0
    Email      string
    KeyVersion string
    TokenType  string  // "access" or "refresh"
}

type TokenPair struct {
    AccessToken  string
    RefreshToken string
    ExpiresIn    int64
}

func NewKeyStore(cfg KeyStoreConfig) (*KeyStore, error)
func (ks *KeyStore) GenerateTokenPair(claims TokenClaims) (*TokenPair, error)
func (ks *KeyStore) ValidateToken(token string) (*TokenClaims, error)
func (ks *KeyStore) AddPreviousKey(version, encodedKey string) error
func IsPASETOToken(token string) bool
```

### Token Structure
- **Format**: `v4.local.<encrypted_payload>`
- **Claims**: `sub`, `email`, `key_version`, `token_type`
- **Access TTL**: 1 hour | **Refresh TTL**: 7 days
- **Encryption**: XChaCha20-Poly1305 + BLAKE2b-MAC

### Configuration

```env
SAVETOINK_AUTH_BACKEND=auth0              # Auth0 identity provider
SAVETOINK_AUTH0_DOMAIN=...                # Auth0 config (required)
SAVETOINK_AUTH0_AUDIENCE=...
SAVETOINK_AUTH0_CLIENT_ID=...
SAVETOINK_AUTH0_CLIENT_SECRET=...
SAVETOINK_PASETO_KEY=<base64-32-bytes>    # PASETO token layer (required)
SAVETOINK_PASETO_KEY_VERSION=v1
```

### API Changes

**`/v1/auth/token` response:**
```json
{
  "access_token": "v4.local.<encrypted>",
  "refresh_token": "v4.local.<encrypted>",
  "token_type": "Bearer",
  "expires_in": 3600,
  "email": "user@example.com"
}
```

**Authorization header** — same pattern, different token:
```
Authorization: Bearer v4.local.<encrypted>
```

---

## Estimated Timeline

| Phase | Status | Key Deliverables |
|-------|--------|------------------|
| Phase 1: Foundation | ✅ Done | PASETO package, 95% coverage |
| Phase 2: Configuration | ✅ Done | Required config, validation |
| Phase 3: Implementation | ✅ Done | Handler, middleware, cleanup |
| Phase 4: Frontend | ⬜ Pending | Frontend updates |
| Phase 5: Deployment | ⬜ Pending | Key generation, deploy |

## Resources

- [PASETO Specification](https://github.com/paseto-standard/paseto-spec)
- [go-paseto Library](https://github.com/aidantwoods/go-paseto)
- `aidanwoods.dev/go-paseto` v1.6.0 — PASETO library (added)
- `github.com/auth0/go-jwt-middleware/v3` — JWT validation (to be removed)
