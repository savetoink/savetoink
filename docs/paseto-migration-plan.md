# PASETO Migration Plan

## Executive Summary

This document outlines the plan to replace JWT with PASETO (Platform-Agnostic SEcurity TOkens) for webapp user authentication. The goal is to improve security by eliminating JWT-related vulnerabilities while maintaining the Auth0 integration for user identity management.

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
- `github.com/auth0/go-jwt-middleware/v3` - JWT validation library

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
- Versioned protocol (v3, v4)
- Clear separation between local and public tokens
- No algorithm confusion
- Standardized security practices

### PASETO Token Types
- **Local tokens (v3.local, v4.local)**: Symmetric encryption - For storing claims securely
- **Public tokens (v3.public, v4.public)**: Asymmetric signatures - For sharing claims externally

## Migration Strategy

### Phase 1: Initial Assessment and Setup (Week 1)

#### Tasks
1. **Research go-paseto library**
   - [ ] Review library documentation: https://github.com/aidantwoods/go-paseto
   - [ ] Understand API surface and capabilities
   - [ ] Verify v4 token support (recommended version)
   - [ ] Check test coverage and maintenance status

2. **Security analysis**
   - [ ] Identify all JWT usage points in codebase
   - [ ] Document current token claims and their purposes
   - [ ] Analyze token lifetime and refresh strategy
   - [ ] Identify security requirements for PASETO keys

3. **Key management design**
   - [ ] Design key rotation strategy
   - [ ] Determine key storage approach (AWS Secrets Manager, KMS, or environment)
   - [ ] Define key generation and distribution process
   - [ ] Document key backup and recovery procedures

#### Deliverables
- Research document with go-paseto findings
- Current JWT usage analysis
- Key management design document

---

### Phase 2: Design and Planning (Week 2)

#### Tasks
1. **Architecture design**
   - [ ] Design PASETO token structure and claims
   - [ ] Define token lifecycle (issuance, validation, rotation)
   - [ ] Design refresh token strategy
   - [ ] Document migration path from JWT to PASETO

2. **API changes**
   - [ ] Update `/v1/auth/token` endpoint to return PASETO tokens
   - [ ] Define backward compatibility strategy (support both JWT and PASETO during migration)
   - [ ] Update frontend integration requirements
   - [ ] Update OpenAPI specification

3. **Configuration changes**
   - [ ] Add new config options for PASETO keys
   - [ ] Update environment variable documentation
   - [ ] Define migration guide for existing deployments

#### Deliverables
- Architecture design document
- API specification updates
- Configuration guide
- Migration roadmap

---

### Phase 3: Implementation (Weeks 3-5)

#### Week 3: Core PASETO Implementation
1. **Add go-paseto dependency**
   ```bash
   cd backend
   go get github.com/aidantwoods/go-paseto
   ```

2. **Create PASETO package**
   - [ ] Create `backend/lib/internal/server/paseto/paseto.go`
   - [ ] Implement token generation
   - [ ] Implement token validation
   - [ ] Implement key management utilities
   - [ ] Add comprehensive unit tests

3. **Update configuration**
   - [ ] Add `PASETOSymmetricKey` to config
   - [ ] Add `PASETOKeyVersion` to config
   - [ ] Add validation for PASETO configuration
   - [ ] Update config documentation

#### Week 4: Token Exchange Implementation
1. **Update `/v1/auth/token` endpoint**
   - [ ] Modify `HandleAuthTokenExchange` to generate PASETO tokens
   - [ ] Keep Auth0 exchange but replace returned token with PASETO
   - [ ] Update `AuthTokenExchangeResponse` to include PASETO tokens
   - [ ] Add backward compatibility flag to return JWT if requested

2. **Update authentication middleware**
   - [ ] Add `pasetoMiddleware` function
   - [ ] Implement PASETO token validation
   - [ ] Update `NewAccountIDMiddleware` to support PASETO backend
   - [ ] Add graceful fallback for JWT tokens during migration

3. **Add new auth backend constant**
   ```go
   const (
       AuthBackendSharedAPIKey  AuthBackend = "shared_api_key"
       AuthBackendAuth0         AuthBackend = "auth0"
       AuthBackendPASETO        AuthBackend = "paseto"  // NEW
   )
   ```

#### Week 5: Testing and Validation
1. **Unit tests**
   - [ ] Test PASETO token generation
   - [ ] Test PASETO token validation
   - [ ] Test token expiration handling
   - [ ] Test key rotation scenarios
   - [ ] Test error handling

2. **Integration tests**
   - [ ] Test complete authentication flow
   - [ ] Test token refresh mechanism
   - [ ] Test concurrent token validation
   - [ ] Test with different key versions

3. **Migration testing**
   - [ ] Test migration from JWT to PASETO
   - [ ] Test backward compatibility
   - [ ] Test token invalidation on logout
   - [ ] Test concurrent access with mixed tokens

---

### Phase 4: Frontend Integration (Week 6)

#### Tasks
1. **Update frontend authentication**
   - [ ] Update token storage to handle PASETO tokens
   - [ ] Update token refresh logic
   - [ ] Update error handling for authentication failures
   - [ ] Add migration logic for existing JWT tokens

2. **Update TypeScript types**
   - [ ] Update `AuthTokenExchangeResponse` type
   - [ ] Add PASETO-specific error types
   - [ ] Update authentication utility functions

3. **Testing**
   - [ ] E2E tests for authentication flow
   - [ ] Browser storage tests
   - [ ] Token refresh tests
   - [ ] Migration tests

---

### Phase 5: Deployment and Monitoring (Week 7)

#### Tasks
1. **Gradual rollout**
   - [ ] Deploy to staging environment
   - [ ] Test with real Auth0 integration
   - [ ] Monitor for authentication errors
   - [ ] Performance testing

2. **Production deployment**
   - [ ] Deploy with feature flag for gradual rollout
   - [ ] Monitor authentication success/failure rates
   - [ ] Monitor token validation latency
   - [ ] Set up alerts for authentication failures

3. **Post-deployment monitoring**
   - [ ] Monitor key rotation process
   - [ ] Track token lifecycle metrics
   - [ ] Monitor for security events
   - [ ] Gather performance metrics

---

### Phase 6: Cleanup and Documentation (Week 8)

#### Tasks
1. **Remove JWT dependencies**
   - [ ] Remove `go-jwt-middleware` dependency
   - [ ] Remove JWT validation code
   - [ ] Clean up unused imports
   - [ ] Remove Auth0-specific token validation

2. **Update documentation**
   - [ ] Update API documentation
   - [ ] Update deployment guide
   - [ ] Update developer guide
   - [ ] Update architecture diagrams

3. **Final testing**
   - [ ] Run full test suite
   - [ ] Security audit
   - [ ] Performance benchmarking
   - [ ] Documentation review

---

## Technical Implementation Details

### PASETO Token Structure

```go
// TokenClaims represents the claims stored in PASETO tokens
type TokenClaims struct {
    Subject    string   `json:"sub"`      // Account ID from Auth0
    Email      string   `json:"email"`    // User email
    IssuedAt   int64    `json:"iat"`      // Issued at timestamp
    ExpiresAt  int64    `json:"exp"`      // Expiration timestamp
    KeyVersion string   `json:"v"`        // Key version identifier
    TokenType  string   `json:"type"`     // "access" or "refresh"
}
```

### Key Management

**Key Storage:**
- Use AWS Secrets Manager for key storage
- Support multiple key versions for rotation
- Key format: Base64-encoded 32-byte key for v4.local tokens

**Key Rotation:**
- Rotate keys every 90 days
- Maintain up to 2 previous key versions for validation
- Graceful transition period: 30 days

**Key Configuration:**
```go
// Config additions
type Config struct {
    // ... existing fields ...

    // PASETO
    PASETOSymmetricKey      string
    PASETOKeyVersion        string
    PASETOAccessTTL         time.Duration
    PASETORefreshTTL        time.Duration
}
```

### Authentication Flow Changes

**Current (JWT):**
```
Frontend → Auth0 → Frontend (auth code)
Frontend → Backend /v1/auth/token (exchange code)
Backend → Auth0 (exchange for JWT)
Auth0 → Backend (JWT tokens)
Backend → Frontend (Auth0 JWT)
Frontend → Backend (Auth0 JWT in Authorization header)
Backend validates JWT using JWKS
```

**New (PASETO):**
```
Frontend → Auth0 → Frontend (auth code)
Frontend → Backend /v1/auth/token (exchange code)
Backend → Auth0 (exchange for JWT)
Auth0 → Backend (JWT tokens)
Backend validates JWT to extract claims
Backend generates PASETO tokens from claims
Backend → Frontend (PASETO tokens)
Frontend → Backend (PASETO token in Authorization header)
Backend validates PASETO token using local key
```

### API Changes

**Updated `/v1/auth/token` Response:**
```json
{
  "access_token": "v4.local.<encrypted_paseto_token>",
  "refresh_token": "v4.local.<encrypted_refresh_token>",
  "token_type": "Bearer",
  "expires_in": 3600,
  "email": "user@example.com"
}
```

**New Headers:**
- `Authorization: Bearer v4.local.<encrypted_paseto_token>`

### Backward Compatibility Strategy

**Transition Period:**
1. Support both JWT and PASETO tokens simultaneously
2. Add `X-Auth-Type` header to indicate token type (optional)
3. Log deprecation warnings for JWT usage
4. Gradually migrate users from JWT to PASETO on token refresh

**Migration Logic:**
```go
func validateToken(token string) (string, error) {
    if strings.HasPrefix(token, "v4.local.") {
        return validatePASETOToken(token)
    } else {
        log.Warn("JWT token detected - migrate to PASETO")
        return validateJWTToken(token)
    }
}
```

---

## Security Considerations

### Key Security
1. **Key Storage**: Use AWS Secrets Manager with encryption
2. **Key Rotation**: Rotate every 90 days with 30-day grace period
3. **Key Access**: Restrict key access to necessary services only
4. **Key Backup**: Secure backup of current and previous keys

### Token Security
1. **Token Lifetime**: Short-lived access tokens (1 hour)
2. **Refresh Tokens**: Longer-lived (7 days)
3. **Token Invalidation**: Implement logout and token revocation
4. **Secure Transmission**: Always use HTTPS

### Attack Mitigation
1. **Replay Attacks**: Include issued-at timestamp and nonce
2. **Timing Attacks**: Use constant-time comparison for token validation
3. **Side-Channel Attacks**: Follow PASETO specification for implementation

---

## Risk Assessment

### High Risk
- **Key loss during rotation**: Mitigated by maintaining multiple key versions
- **Breaking changes during migration**: Mitigated by backward compatibility period
- **Performance degradation**: Mitigated by local token validation (no JWKS fetch)

### Medium Risk
- **Complexity of migration**: Mitigated by phased approach and testing
- **Frontend integration issues**: Mitigated by comprehensive testing and gradual rollout
- **Key management overhead**: Mitigated by using AWS Secrets Manager

### Low Risk
- **go-paseto library issues**: Mitigated by choosing well-maintained library
- **Configuration errors**: Mitigated by validation and documentation
- **Deployment issues**: Mitigated by staging testing and gradual rollout

---

## Testing Strategy

### Unit Tests
- Token generation and validation
- Key rotation scenarios
- Error handling
- Token expiration

### Integration Tests
- Complete authentication flow
- Token refresh mechanism
- Concurrent token validation
- Migration scenarios

### End-to-End Tests
- Frontend authentication flow
- Token refresh in browser
- Session management
- Logout and token invalidation

### Security Tests
- Token manipulation attempts
- Replay attack prevention
- Key compromise scenarios
- Timing attack prevention

### Performance Tests
- Token validation latency
- Concurrent token validation
- Memory usage
- Key rotation performance impact

---

## Rollback Plan

If critical issues arise during migration:

1. **Immediate Rollback**
   - Switch to JWT backend via config change
   - No code changes required

2. **Gradual Rollback**
   - Reduce PASETO adoption rate
   - Increase JWT usage
   - Investigate and fix issues

3. **Monitoring During Rollback**
   - Monitor authentication success rates
   - Track error rates
   - Review security logs

---

## Success Metrics

### Security
- No JWT-related vulnerabilities
- Secure key management implemented
- Proper token lifecycle management

### Performance
- Token validation latency < 10ms
- No JWKS fetch overhead
- Reduced memory footprint

### Reliability
- Authentication success rate > 99.9%
- Minimal downtime during migration
- Zero data loss

### Operations
- Simplified key management
- Reduced external dependencies
- Clear documentation

---

## Estimated Timeline

| Phase | Duration | Key Deliverables |
|-------|----------|------------------|
| Phase 1: Initial Assessment | 1 week | Research and analysis documents |
| Phase 2: Design and Planning | 1 week | Architecture design, API specs |
| Phase 3: Implementation | 3 weeks | PASETO implementation, tests |
| Phase 4: Frontend Integration | 1 week | Updated frontend, tests |
| Phase 5: Deployment | 1 week | Staging and production deployment |
| Phase 6: Cleanup | 1 week | Documentation, cleanup |
| **Total** | **8 weeks** | Complete migration |

---

## Resources

### References
- [PASETO Specification](https://github.com/paseto-standard/paseto-spec)
- [go-paseto Library](https://github.com/aidantwoods/go-paseto)
- [JWT vs PASETO](https://paragonie.com/blog/2017/03/jwt-json-web-tokens-is-bad-standard-that-everyone-should-avoid)

### Dependencies
- `github.com/aidantwoods/go-paseto` - PASETO library
- AWS Secrets Manager - Key storage

---

## Conclusion

This migration plan provides a comprehensive approach to replacing JWT with PASETO for improved security. The phased approach minimizes risk while maintaining backward compatibility. The timeline is realistic with built-in testing and validation at each phase.

The key benefits of this migration include:
- Enhanced security through PASETO's secure-by-design approach
- Simplified token validation (no JWKS fetch)
- Better control over token lifecycle
- Reduced external dependencies

Success depends on thorough testing, careful key management, and gradual deployment with monitoring.
