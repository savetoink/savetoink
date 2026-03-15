# Move `auth` Package to Internal - Implementation Plan

**Package:** `backend/lib/auth` → `backend/lib/internal/auth`

**Status:** ✅ Analysis Complete - Ready for Implementation

**Date:** March 15, 2026

---

## Package Analysis

### Why `auth` is an excellent candidate:

✅ **Zero entry point dependencies** - No `cmd/*` or `cli/*` files import it  
✅ **Zero external dependencies** - Only imports standard library (`context`, `errors`)  
✅ **Simple, focused API** - 4 pure context helper functions  
✅ **Internal-only usage** - Only used by other `backend/lib/*` packages  
✅ **Well-tested** - Includes comprehensive test coverage  
✅ **Low risk** - Minimal complexity, easy to verify  

### Current Package Structure

```
backend/lib/auth/
├── context.go         # Context helper functions
└── context_test.go    # Tests for context functions
```

### Exported API

**Functions (4):**
- `GetAccountIDFromCtx(ctx context.Context) string`
- `GetAuthErrorFromCtx(ctx context.Context) error`
- `AddAccountIDToCtx(ctx context.Context, accountID string) context.Context`
- `AddAuthErrorToCtx(ctx context.Context, msg string) context.Context`

**Dependencies:** `context`, `errors` (standard library only)

### Import Analysis

**Consumers (10 files total):**

**Main packages (6 files):**
- `backend/lib/server/auth/auth.go`
- `backend/lib/logging/middleware.go`
- `backend/lib/logging/record.go`
- `backend/lib/server/handlers/user_profile.go`
- `backend/lib/server/handlers/sends.go`
- `backend/lib/server/handlers/article.go`

**Test files (4 files):**
- `backend/lib/server/handlers/user_profile_test.go`
- `backend/lib/server/handlers/sends_test.go`
- `backend/lib/server/handlers/article_test.go`
- `backend/lib/server/auth/auth_test.go`

**Entry points:** None (no direct imports by `cmd/*` or `cli/*`)

---

## Implementation Steps

### Step 1: Create internal directory

```bash
mkdir -p backend/lib/internal/auth
```

**Expected result:** New empty directory created

---

### Step 2: Copy files to internal

```bash
cp backend/lib/auth/*.go backend/lib/internal/auth/
```

**Files to copy:**
- `context.go`
- `context_test.go`

**Expected result:** Both files copied to `backend/lib/internal/auth/`

---

### Step 3: Update imports in consumers

**Pattern to apply:**
```go
// Change from:
"github.com/shaftoe/savetoink/backend/lib/auth"

// Change to:
"github.com/shaftoe/savetoink/backend/lib/internal/auth"
```

**Files to update (10 total):**

#### Main packages (6 files):
1. `backend/lib/server/auth/auth.go`
2. `backend/lib/logging/middleware.go`
3. `backend/lib/logging/record.go`
4. `backend/lib/server/handlers/user_profile.go`
5. `backend/lib/server/handlers/sends.go`
6. `backend/lib/server/handlers/article.go`

#### Test files (4 files):
7. `backend/lib/server/handlers/user_profile_test.go`
8. `backend/lib/server/handlers/sends_test.go`
9. `backend/lib/server/handlers/article_test.go`
10. `backend/lib/server/auth/auth_test.go`

**Implementation command:**
```bash
cd backend/lib && sed -i '' 's|github.com/shaftoe/savetoink/backend/lib/auth|github.com/shaftoe/savetoink/backend/lib/internal/auth|g' \
  server/auth/auth.go \
  logging/middleware.go \
  logging/record.go \
  server/handlers/user_profile.go \
  server/handlers/sends.go \
  server/handlers/article.go \
  server/handlers/user_profile_test.go \
  server/handlers/sends_test.go \
  server/handlers/article_test.go \
  server/auth/auth_test.go
```

**Expected result:** All 10 files updated with new import path

---

### Step 4: Verify all packages build

```bash
cd backend/lib && go build ./...
```

**Expected result:** All packages compile successfully with no errors

---

### Step 5: Run tests

```bash
cd backend && go test ./lib/... -short
```

**Expected result:** All tests pass (including the copied tests in `internal/auth/`)

**Key tests to verify:**
- `backend/lib/internal/auth/context_test.go` - All tests pass
- `backend/lib/server/auth/auth_test.go` - All tests pass (uses internal/auth)
- `backend/lib/server/handlers/*_test.go` - All tests pass

---

### Step 6: Fix formatting

```bash
cd backend/lib && gofmt -w internal/auth/*.go
```

**Expected result:** All files properly formatted

**Also verify updated consumer files:**
```bash
cd backend/lib && gofmt -w server/auth/auth.go logging/*.go server/handlers/*.go
```

---

### Step 7: Run linting

```bash
just check
```

**Expected result:** Linting passes with no errors

**Note:** Ignore pre-existing issues in `backend/lib/internal/task/runner.go` (known slog formatting issue)

---

### Step 8: Remove old directory

```bash
rm -rf backend/lib/auth
```

**Expected result:** Old `backend/lib/auth/` directory removed

---

### Step 9: Verify all entry points build

#### HTTP Server
```bash
go build -buildvcs=false -o /tmp/savetoink-http ./backend/cmd/http/
```

**Expected result:** ✓ HTTP server builds

#### Lambda API
```bash
go build -buildvcs=false -o /tmp/savetoink-lambda ./backend/cmd/lambda/
```

**Expected result:** ✓ Lambda API builds

#### Lambda Processor
```bash
go build -buildvcs=false -o /tmp/savetoink-lambda-processor ./backend/cmd/lambda/processor/
```

**Expected result:** ✓ Lambda processor builds

#### Lambda Scheduler
```bash
go build -buildvcs=false -o /tmp/savetoink-lambda-scheduler ./backend/cmd/lambda/scheduler/
```

**Expected result:** ✓ Lambda scheduler builds

#### CLI
```bash
go build -buildvcs=false -o /tmp/savetoink-cli ./cli/savetoink/
```

**Expected result:** ✓ CLI builds

---

### Step 10: Verify no old references

```bash
grep -r "lib/auth" backend/cmd/ cli/savetoink/ backend/lib/ 2>/dev/null | grep -v internal
```

**Expected result:** No output (no references to old `lib/auth` path remain)

---

## Success Criteria

- [ ] All packages build successfully
- [ ] All tests pass
- [ ] Linting passes (excluding known issues)
- [ ] All 5 entry points build
- [ ] Old `backend/lib/auth/` directory removed
- [ ] No references to `lib/auth` remain
- [ ] All 10 consumer files updated correctly

---

## Risk Assessment: **Very Low**

### Risk Factors

| Factor | Risk Level | Details |
|---------|-------------|---------|
| **Import complexity** | Very Low | Only standard library imports |
| **Consumer count** | Low | Only 10 files to update |
| **Test complexity** | Low | Simple context functions, easy to test |
| **Entry point impact** | None | No entry points import it directly |
| **Dependencies** | None | Zero external or internal dependencies |

### Rollback Plan

If issues arise during implementation:

1. **Revert import changes:**
   ```bash
   cd backend/lib && sed -i '' 's|github.com/shaftoe/savetoink/backend/lib/internal/auth|github.com/shaftoe/savetoink/backend/lib/auth|g' \
     server/auth/auth.go \
     logging/middleware.go \
     logging/record.go \
     server/handlers/*.go
   ```

2. **Restore from backup:**
   ```bash
   rm -rf backend/lib/internal/auth
   mkdir -p backend/lib/auth
   # Restore files from git or backup
   ```

---

## Estimated Time: 15-20 minutes

- Step 1-2 (Setup): 2 minutes
- Step 3 (Import updates): 5 minutes
- Step 4-5 (Testing): 8 minutes
- Step 6-7 (Linting): 3 minutes
- Step 8-10 (Cleanup): 2 minutes

---

## Benefits of Moving `auth` to Internal

1. **Clearer API Boundaries:** Explicitly separates internal context utilities from public API
2. **Easier Refactoring:** Can change context helper implementation without external concerns
3. **Reduced Coupling:** Prevents external consumers from depending on internal auth context details
4. **Consistent Structure:** Follows established pattern (`internal/task`, `internal/apperrors`)
5. **Go Best Practices:** Aligns with Go's convention for `internal/` packages

---

## Post-Implementation Verification

After completion, verify:

### Directory structure:
```
backend/lib/
├── internal/
│   ├── auth/
│   │   ├── context.go
│   │   └── context_test.go
│   ├── apperrors/
│   │   ├── errors.go
│   │   └── errors_test.go
│   └── task/
│       └── (existing task package)
└── (other packages)
```

### Import usage:
```bash
grep -r "internal/auth" backend/lib/ | wc -l
# Should show 10 files using new import
```

### Test coverage:
```bash
cd backend/lib/internal/auth && go test -cover
# Should show good test coverage (>80%)
```

---

## Next Steps

After successful completion of `auth` package move:

1. Update `.opencode/plans/backend-internal-restructuring.md` with completion status
2. Consider next packages to move:
   - `validation` (4 import updates, simple dependencies)
   - `processor` (used only by server package)
   - `repository` (larger package, more dependencies)

---

## Notes

- The `auth` package contains pure context helper functions with no side effects
- No configuration or external service dependencies
- All functions are pure helpers that manipulate Go's `context.Context`
- This is the second internal package move after `apperrors`
- Establishes pattern for moving context utilities to internal
