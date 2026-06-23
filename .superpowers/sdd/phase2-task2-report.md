# Phase 2 Task 2 Report

## Scope

Executed `.superpowers/sdd/phase2-task2-brief.md`: add JWT authentication middleware for the API.

## Changes

- Added `apps/api/internal/auth/middleware_test.go`.
- Added tests for missing bearer token, invalid bearer token, and a valid HS256 token containing `sub`, `aud=authenticated`, and a future `exp`.
- Added `apps/api/internal/auth/middleware.go`.
- Implemented `RequireUser(jwtSecret string)` to validate `Authorization: Bearer ...`, accept HS256 JWTs, parse `sub` as a UUID, and store the user ID in the request context.
- Implemented `UserIDFromContext(ctx context.Context)`.
- Added Go module dependencies:
  - `github.com/golang-jwt/jwt/v5 v5.3.1`
  - `github.com/google/uuid v1.6.0`

## Verification

Ran RED test:

```bash
go test ./apps/api/internal/auth -v
```

Result:

- Exit code: `1`
- Failed because `RequireUser` and `UserIDFromContext` were undefined.

Ran dependency install:

```bash
cd apps/api
go get github.com/golang-jwt/jwt/v5 github.com/google/uuid
```

Result:

- First attempt failed because `proxy.golang.org` timed out.
- Retried with `GOPROXY=https://goproxy.cn,direct`; exit code: `0`.

Ran GREEN test:

```bash
cd apps/api
go test ./internal/auth -v
```

Result:

- Exit code: `0`
- `TestRequireUserRejectsMissingBearerToken`: PASS
- `TestRequireUserRejectsInvalidBearerToken`: PASS
- `TestRequireUserStoresValidTokenSubjectInRequestContext`: PASS

Ran final API test suite:

```bash
cd apps/api
go test ./...
```

Result:

- Exit code: `0`
- `cmd/api`, `internal/config`, and `internal/db`: no test files.
- `internal/auth`: PASS.
- `internal/http`: PASS.

## Notes

- The test JWT helper signs HS256 tokens with standard library HMAC code so the RED phase fails on missing middleware symbols, not on missing test-only dependencies.
