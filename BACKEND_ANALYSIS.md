# Backend Code Quality Analysis: backend/internal

**Generated:** March 6, 2026
**Scope:** All Go files in `backend/internal/` (50 files analyzed)

---

## Executive Summary

This analysis identified **61 issues** across the backend codebase:
- **5 Critical** issues requiring immediate attention
- **11 High** priority issues
- **20 Medium** priority issues
- **25 Low** priority issues

**Major Areas of Concern:**
1. Testing - Repository and handler layers have minimal test coverage
2. Service Design - God object with 19 methods violating SOLID principles
3. Security - Webhook authentication uses query parameters instead of HMAC signatures
4. Coupling - Type assertions and tight coupling between layers
5. Error Handling - Inconsistent patterns throughout

---

## 1. Code Smells

### 1.1 Long Functions/Methods

| File | Function | Lines | Severity | Description |
|------|----------|-------|----------|-------------|
| `handlers_auth.go` | `handleAuthTokenExchange` | 25-91 (67 lines) | Medium | Complex auth token exchange with multiple concerns |
| `article_query.go` | `GetMetadataByAccount` | 87-137 (51 lines) | Medium | Pagination logic with offset calculation loop |
| `router.go` | `setupLogging` | 108-154 (47 lines) | Medium | Logging setup with multiple conditional branches |
| `extractor.go` | `ExtractFromURL` | 35-71 (37 lines) | Low | Moderate complexity but acceptable |
| `service/article_creation.go` | `CreateArticle` | 20-65 (46 lines) | Low | Orchestrates multiple operations, acceptable for use case |

**Recommendation:** Break down `handleAuthTokenExchange` into smaller functions: `validateTokenRequest`, `exchangeToken`, `extractEmailFromToken`, `storeUserEmail`. Break `GetMetadataByAccount` into `validatePagination`, `skipToPage`, `fetchPageData`.

### 1.2 Complex Conditional Logic

| File | Lines | Severity | Description |
|------|-------|----------|-------------|
| `auth/bounce.go` | 44-59 | Medium | Nested conditionals in `checkDeviceEmail` make flow hard to follow |
| `auth/bounce.go` | 62-83 | Medium | Complex logic in `handleBouncingEmail` with multiple returns |
| `article_query.go` | 114-124 | Low | Loop to skip pagination offset could be clearer |
| `handlers_article.go` | 178-220 | Low | Multiple conditional checks in `handleSendArticle` |

**Recommendation:** Refactor `checkDeviceEmail` and `handleBouncingEmail` to use early returns and reduce nesting. Consider using a struct to represent device email validation result.

### 1.3 Duplicate Code

| File | Lines | Severity | Description |
|------|-------|----------|-------------|
| `handlers_article.go` | 19-28, 55-64, 101-113 | High | Repeated request validation patterns (JSON decode, error handling) |
| `repository/dynamodb/*.go` | Various | High | Unmarshal error handling pattern repeated 8+ times |
| `handlers_user_profile.go` | 52-60, 73-82 | Medium | Similar error handling for device operations |
| `service/user_profile.go` | 36-60, 63-96 | Medium | Duplicate profile get/update logic |
| `handlers_auth.go` | 144-168, 187-211 | Medium | JWT token parsing logic duplicated between email and subject extraction |

**Recommendation:** Create helper functions:
- `decodeAndValidateRequest(w, r, req)` for common handler validation
- `unmarshalDynamoDBItem(item, target)` for repository operations
- `parseJWTClaims(token, target)` for JWT parsing
- Consider a validation middleware or request struct methods

### 1.4 Poor Naming Conventions

| File | Issue | Severity | Description |
|------|--------|----------|-------------|
| `article_creation.go` | `eg` variable (line 31) | Low | Short variable name, could be `errorGroup` |
| `article_send.go` | `send` variable (line 41) | Low | Variable noun vs method verb confusion, rename to `sendRecord` |
| `handlers_article.go` | `h` parameter throughout | Low | Standard but could be more explicit as `handlers` |
| `middleware.go` | `lc` variable (line 52) | Low | `lambdaContext` would be clearer |

**Recommendation:** Use more descriptive names for clarity, especially for variables with longer lifespans.

### 1.5 Magic Numbers/Strings

| File | Lines | Severity | Description |
|------|-------|----------|-------------|
| `content/extractor.go` | 175 | Low | `len(parts) < 2` - should use named constant |
| `content/extractor.go` | 146 | Low | `len(parts) != jwtPartsCount` - good use of constant |
| `content/extractor.go` | 227 | Low | Hardcoded `<p>`, `</p>` replacements in `stripHTML` |
| `middleware.go` | 20060102-150405.000 | Low | Time format string should be constant |
| `handlers_auth.go` | 21-22 | Low | `jwtPartsCount = 3` - good use of constant |

**Recommendation:**
```go
const (
    titlePartsMinCount = 2
    htmlParagraphStart = "<p>"
    htmlParagraphEnd   = "</p>"
    requestIDTimeFormat = "20060102-150405.000"
)
```

### 1.6 God Objects/Classes

| File | Severity | Description |
|------|----------|-------------|
| `service/service.go` | High | `Service` struct has too many responsibilities: article CRUD, user profile, sends, bounce handling, email validation. Interface has 19 methods. |
| `server/handlers.go` | Medium | `handlers` struct manages all HTTP concerns |

**Recommendation:** Split `Service` into:
- `ArticleService` (Create, Get, List, Delete, ToggleFavorite)
- `UserProfileService` (GetProfile, SetEmail, SetDeviceEmail)
- `SendService` (SendArticle, CountSends, HandleBounce)
- Use composition in a higher-level service

### 1.7 Feature Envy

| File | Lines | Severity | Description |
|------|-------|----------|-------------|
| `handlers_article.go` | 77-89 | Low | Handler directly formats favoriteFilter, could be service concern |
| `middleware.go` | 154-172 | Medium | Logging middleware knows too much about HTTP request details |

### 1.8 Inappropriate Intimacy

| File | Lines | Severity | Description |
|------|-------|----------|-------------|
| `service/service.go` | 78-81 | Medium | Type assertions to convert repository types break abstraction |
| `handlers_auth.go` | 175-211 | Medium | Direct JWT parsing logic that should be in auth package |

---

## 2. Design Issues

### 2.1 Tight Coupling

| File | Lines | Severity | Description |
|------|-------|----------|-------------|
| `service/service.go` | 78-81 | High | Type assertion `repo.(repository.UserProfileRepository)` couples service to concrete implementation |
| `server/router.go` | 36-42 | Medium | Direct service instantiation creates tight coupling |
| `server/auth/auth.go` | 144 | Medium | `next.ServeHTTP(w, r)` without auth error in Auth0 middleware path |

**Recommendation:** Use dependency injection, pass all required repositories as parameters to `New()`, avoid type assertions.

### 2.2 Low Cohesion

| File | Severity | Description |
|------|----------|-------------|
| `service/bounce.go` | Medium | Bounce handling mixed with user profile concerns |
| `handlers.go` | Medium | Single handlers file contains all HTTP concerns |

### 2.3 Violation of SOLID Principles

| Principle | Violations | Severity |
|-----------|-------------|----------|
| **Single Responsibility** | `Service` class handles articles, profiles, sends, bounces, validation | High |
| **Open/Closed** | `NewAccountIDMiddleware` uses switch statement (lines 38-47) | Medium |
| **Liskov Substitution** | Type assertions in service.go may violate LSP | High |
| **Interface Segregation** | `Service.Interface` has 19 methods, clients depend on all | Medium |
| **Dependency Inversion** | Service depends on concrete repositories via type assertions | High |

**Recommendation:**
- Apply Interface Segregation: Create smaller interfaces like `ArticleReader`, `ArticleWriter`
- Use dependency injection properly
- Consider strategy pattern for auth backends instead of switch

### 2.4 Poor Separation of Concerns

| File | Severity | Description |
|------|----------|-------------|
| `service/article_creation.go` | Medium | Background DB operations mixed with business logic |
| `handlers_article.go` | Medium | Request validation, auth, business logic all in handlers |
| `auth/subscription.go` | Medium | Free tier limit checking mixed with subscription logic |

### 2.5 God Packages

| Package | Severity | Description |
|---------|----------|-------------|
| `service` | Medium | Contains business logic, repository access, email operations, bounce handling |
| `server` | Medium | Contains handlers, middleware, routing, auth concerns |

### 2.6 Circular Dependencies

**Status:** None detected ✅

The codebase has good package structure, no circular imports found.

---

## 3. Security Concerns

### 3.1 SQL Injection Risks

**Status:** N/A ✅

Using DynamoDB, not SQL, so SQL injection not applicable.

### 3.2 Missing Input Validation

| File | Lines | Severity | Description |
|------|-------|----------|-------------|
| `handlers_article.go` | 19-29 | High | Only checks URL is present, doesn't validate URL format |
| `service/user_profile.go` | 149-168 | High | Device email validation only checks domain suffix, doesn't verify email format beyond basic parse |
| `handlers_webhook.go` | 65-76 | Medium | Webhook secret comparison is simple string compare, should use constant-time compare |
| `content/extractor.go` | 232-250 | Medium | URL validation has basic checks but could be more thorough |

**Recommendation:**
```go
// Use crypto/subtle.ConstantTimeCompare for webhook secrets
import "crypto/subtle"
if subtle.ConstantTimeCompare([]byte(secretQueryParam), []byte(h.cfg.MailjetWebhookSecret)) != 1 {
    return errors.New("invalid webhook secret")
}

// Add URL validation
func validateArticleURL(url string) error {
    u, err := url.Parse(url)
    if err != nil {
        return err
    }
    if u.Scheme != "http" && u.Scheme != "https" {
        return errors.New("invalid URL scheme")
    }
    // Add blocklist for private/internal networks
    if isPrivateIP(u.Host) {
        return errors.New("private URLs not allowed")
    }
    return nil
}
```

### 3.3 Hardcoded Secrets/Credentials

| File | Lines | Severity | Description |
|------|-------|----------|-------------|
| `auth/auth.go` | 23 | Low | `adminAccountID = "admin"` is hardcoded but from config so acceptable |
| `auth/subscription.go` | 31 | Low | Free tier limits are constants, acceptable |

### 3.4 Insecure Error Handling

| File | Lines | Severity | Description |
|------|-------|----------|-------------|
| `handlers_article.go` | 36 | Medium | Error details exposed to client: `err.Error()` |
| `handlers_article.go` | 84 | Medium | Database errors propagated to client |
| `handlers_webhook.go` | 36-45 | Medium | Webhook returns 200 before processing, could mask errors |

**Recommendation:** Don't expose internal error details to clients. Use generic error messages and log detailed errors internally.

### 3.5 Missing Authentication/Authorization Checks

| File | Lines | Severity | Description |
|------|-------|----------|-------------|
| `router.go` | 67-105 | Low | Good use of `EnsureAutheticatedMiddleware` on routes |
| `handlers_webhook.go` | 34-63 | Medium | Webhook endpoint uses query param for auth instead of headers/signature |
| `auth/subscription.go` | 22-53 | Low | Free tier checking works, but rate limiting could be added |

**Recommendation:** Use HMAC signature verification for webhooks instead of query parameter secrets.

---

## 4. Performance Issues

### 4.1 N+1 Query Problems

| File | Lines | Severity | Description |
|------|-------|----------|-------------|
| `article_query.go` | 97-100 | Medium | `totalCountByAccount` queries separately from actual data fetch |
| `article_query.go` | 114-124 | Medium | Loop to skip pagination offset makes multiple queries |

**Recommendation:** Consider caching pagination results or using more efficient pagination strategy.

### 4.2 Missing Indexes

**Status:** N/A ✅

Using DynamoDB with GSIs defined in consts/storage.go:
- AccountCreatedAtIndex
- ArticleIdIndex
- AccountSentAtIndex
- DeviceEmailIndex

### 4.3 Inefficient Algorithms

| File | Lines | Severity | Description |
|------|-------|----------|-------------|
| `article_query.go` | 114-124 | Medium | Sequential offset skipping with multiple queries - inefficient for deep pagination |
| `article_deletion.go` | 111-146 | Medium | Fetches all articles before deleting in batches |
| `content/extractor.go` | 196-224 | Medium | `stripHTML` uses manual character-by-character iteration |

**Recommendation:**
```go
// Use cursor-based pagination instead of offset
type PaginationToken struct {
    LastEvaluatedKey map[string]types.AttributeValue
}

// Use regex or html parser for stripHTML
func stripHTML(html string) string {
    re := regexp.MustCompile(`<[^>]*>`)
    return re.ReplaceAllString(html, " ")
}
```

### 4.4 Memory Leaks

| File | Lines | Severity | Description |
|------|-------|----------|-------------|
| `service/article_creation.go` | 31-35 | Medium | Background goroutine with channel, potential goroutine leak if context cancelled |
| `middleware.go` | 42-45 | Low | Body close error logged but not critical |

**Recommendation:** Ensure background goroutine is properly cleaned up on context cancellation.

### 4.5 Unnecessary Database Calls

| File | Lines | Severity | Description |
|------|-------|----------|-------------|
| `article_deletion.go` | 111-146 | Medium | Calls `GetMetadataByAccount` just to count/delete, could use query directly |
| `article_retrieval.go` | 28-34 | Low | Separate call to check article exists before delete - could use conditional delete |

### 4.6 Missing Connection Pooling

**Status:** N/A ✅

Using AWS SDK v2 which manages connection pooling automatically.

---

## 5. Maintainability Issues

### 5.1 Missing or Poor Error Handling

| File | Lines | Severity | Description |
|------|-------|----------|-------------|
| `service/article_creation.go` | 46-49 | Medium | Error stored in article but still returned nil - inconsistent |
| `handlers_article.go` | 25-28 | Low | Simple validation but no logging of failed requests |
| `repository/dynamodb/send.go` | 115-120 | Low | Generic error "failed to count sends" loses context |

### 5.2 Inconsistent Error Handling Patterns

| Scope | Severity | Description |
|-------|----------|-------------|
| Multiple files | Medium | Some return `errors.New()`, some return `fmt.Errorf()`, some return wrapped errors |
| Handlers | Medium | Some log errors, some don't. Some set 500, some set 400 |

**Recommendation:** Establish error handling conventions:
```go
// Use sentinel errors for common cases
var (
    ErrNotFound = errors.New("not found")
    ErrInvalid = errors.New("invalid input")
)

// Use fmt.Errorf with %w for wrapping
return fmt.Errorf("failed to create article: %w", err)
```

### 5.3 Missing Logging

| File | Severity | Description |
|------|----------|-------------|
| `service/bounce.go` | Medium | No logging of bounce handling events |
| `service/article_send.go` | Low | Logs send update failure but not success |
| `handlers_webhook.go` | Low | Webhook processing errors added to context but not logged |

### 5.4 Poor Code Organization

| File | Severity | Description |
|------|----------|-------------|
| `service` package | Medium | 7 files, concerns somewhat scattered |
| `server` package | Medium | 7 files, handlers and auth mixed |

### 5.5 Lack of Documentation

| File | Severity | Description |
|------|----------|-------------|
| `service/*.go` | Low | Good function-level comments present |
| `repository/dynamodb/*.go` | Low | Most methods have comments |
| `model/*.go` | Medium | Some structs lack field documentation |

**Recommendation:** Add package-level documentation explaining architecture and data flow.

### 5.6 Inconsistent Patterns Across Files

| Pattern | Inconsistency | Severity |
|---------|---------------|----------|
| Error handling | Mix of `errors.New`, `fmt.Errorf`, wrapping | Medium |
| Context usage | Some functions use context, some don't | Low |
| Response encoding | Some use `json.NewEncoder(w).Encode`, some use `writeJSON` helper | Low |
| Variable naming | Mix of short and descriptive names | Low |

---

## 6. Testing Issues

### 6.1 Missing Tests

| File | Severity | Description |
|------|----------|-------------|
| `config/config.go` | High | No tests for configuration loading and validation |
| `repository/dynamodb/article.go` | High | No tests for repository implementation |
| `repository/dynamodb/profile.go` | High | No tests for profile repository |
| `repository/dynamodb/send.go` | High | No tests for send repository |
| `service/article_creation.go` | Medium | No integration tests for article creation flow |
| `service/article_send.go` | Medium | No tests for send operations |
| `service/bounce.go` | Medium | No tests for bounce handling |
| `service/user_profile.go` | Medium | No tests for profile operations |
| `handlers/*.go` (except auth) | Medium | Missing handler tests |
| `server/middleware.go` | Medium | No tests for CORS, logging, request ID middleware |

### 6.2 Hard-to-Test Code

| File | Lines | Severity | Description |
|------|-------|----------|-------------|
| `service/service.go` | 68-91 | High | `New()` function creates concrete dependencies - hard to mock |
| `service/article_creation.go` | 72-94 | High | Background goroutines make testing difficult |
| `handlers_article.go` | Various | Medium | Tightly coupled to service, no interface injection |
| `middleware.go` | 74-94 | Medium | Logging middleware hard to test due to slog coupling |

**Recommendation:** Use dependency injection and interfaces to make code testable:
```go
type Dependencies struct {
    Extractor   *content.Extractor
    Generator   *epub.Generator
    Sender      email.Sender
    Repo        repository.ArticlesRepository
    UserProfile repository.UserProfileRepository
    SendsRepo   repository.SendsRepository
}

func New(deps Dependencies) *Service {
    return &Service{
        extractor:       deps.Extractor,
        generator:       deps.Generator,
        sender:          deps.Sender,
        repo:            deps.Repo,
        userProfileRepo: deps.UserProfile,
        sendsRepo:       deps.SendsRepo,
    }
}
```

### 6.3 Test Smells

| Test File | Issue | Severity | Description |
|-----------|-------|----------|-------------|
| `service/service_test.go` | 104-217 | Medium | Mock repository with complex sorting/pagination logic - test implementation details |
| `handlers_auth_test.go` | 92-105 | Low | Creates test server but doesn't cleanup properly |
| `email/mailjet/sender_test.go` | 160-281 | Low | Repetitive test table structure |

---

## Summary by Severity

### Critical Issues (5)

1. **Service interface too large (19 methods)** - Violates SRP and ISP
2. **Type assertions in service.go** - Breaks abstraction, creates tight coupling
3. **Missing repository tests** - Critical data layer untested
4. **Missing handler tests** - API endpoints untested
5. **No webhook HMAC verification** - Security vulnerability

### High Issues (11)

1. Duplicate error handling patterns across handlers and repositories
2. Complex conditional logic in auth/bounce.go
3. Insecure webhook secret verification (query param vs signature)
4. Missing input validation for article URLs
5. Background goroutine in article creation - potential leaks
6. Inefficient pagination with multiple queries
7. Device email validation insufficient
8. Inconsistent error handling throughout codebase
9. No configuration tests
10. No service layer tests (article operations, bounce, profile)
11. Hard-to-test code due to lack of dependency injection

### Medium Issues (20)

1. Long functions in handlers and repository
2. God package concerns in service and server
3. Poor separation of concerns in some modules
4. SSRF risk in URL fetching (noted but acceptable)
5. Unmarshal error handling repeated 8+ times
6. Webhook returns 200 before processing errors
7. Duplicate profile get/update logic
8. JWT parsing duplicated in auth handlers
9. Pagination offset skipping loop
10. Manual HTML stripping inefficient
11. Error details exposed to clients
12. Missing bounce handling logging
13. Test mocks too complex (testing implementation)
14. CORS middleware allows any origin
15. No rate limiting on API endpoints
16. Repository batch deletion fetches all first
17. Missing logging in several critical paths
18. Inconsistent response encoding patterns
19. Poor variable naming in some contexts
20. Missing package-level documentation

### Low Issues (25)

1. Magic numbers/strings without constants
2. Short variable names (eg, lc, etc.)
3. Missing struct field documentation
4. Minor code organization improvements
5. Could extract common validation logic
6. Some error messages generic
7. Request ID generation could use UUID
8. Some functions could use early returns
9. Minor inconsistency in naming conventions
10. Could add more constants for HTML tags
11. Body close warnings logged but not critical
12. Default email values in responses
13. Some test tables could be more concise
14. Test servers may not cleanup properly
15. Minor: hardcoded user agent list
16. Could use constant for pagination logic
17. Minor: file permissions magic number
18. Some context usage inconsistent
19. Could extract common response helpers
20. Minor: duplicate import cleanup needed
21. Could add more struct validation tags
22. Minor: error variable naming inconsistency
23. Some comments could be more detailed
24. Could add request struct methods for validation
25. Minor: could extract HTTP client configuration

---

## Recommended Prioritized Improvements

### Phase 1: Critical & High Priority

1. **Add comprehensive tests for repository layer** - Critical data layer currently untested
2. **Add handler tests for all API endpoints** - API endpoints lack test coverage
3. **Implement HMAC webhook verification** - Security vulnerability with query param auth - **FIXED** (we accept the risk)
4. **Refactor Service to use dependency injection** - Remove type assertions, improve testability **DONE**
5. **Split Service into smaller focused services** - Apply SRP, reduce interface size from 19 methods **DONE**
6. **Add input validation for all user inputs** - URL validation, email validation, webhook secrets **DONE**
7. **Extract common error handling patterns** - Reduce code duplication, improve consistency

### Phase 2: Medium Priority

8. **Implement HMAC signature verification for webhooks** - Replace query param auth with HMAC
9. **Add configuration tests** - Validate config loading and environment variable handling
10. **Implement cursor-based pagination** - Replace inefficient offset-based pagination
11. **Add comprehensive logging for critical paths** - Bounce handling, sends, errors
12. **Improve error handling consistency** - Establish conventions, use sentinel errors
13. **Add rate limiting to API** - Protect against abuse
14. **Implement better background goroutine handling** - Ensure proper cleanup, prevent leaks
15. **Add integration tests** - Test end-to-end flows

### Phase 3: Low Priority

16. **Extract magic numbers/strings to constants** - Improve code readability
17. **Improve variable naming where needed** - Use descriptive names
18. **Add more documentation** - Package-level docs, struct field docs
19. **Optimize HTML stripping** - Use regex or HTML parser
20. **Minor refactoring for consistency** - Align patterns across codebase

---

## Positive Findings

Despite the issues above, the codebase has many strengths:

- ✅ Clean package structure with no circular dependencies
- ✅ Good use of interfaces for repository abstractions
- ✅ Consistent use of context throughout
- ✅ Comprehensive OpenAPI specification maintained
- ✅ Good separation between handlers, services, and repositories
- ✅ Proper use of Go idioms (defer, error wrapping)
- ✅ Well-organized constants and configurations
- ✅ Good test coverage for authentication
- ✅ Proper use of middleware for cross-cutting concerns
- ✅ Type-safe email validation with domain checking
