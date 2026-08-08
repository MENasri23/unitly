# Refresh Token Endpoint Implementation Summary

## Overview

Implemented a complete token refresh endpoint that follows security best practices and integrates seamlessly with the existing authentication system.

## Endpoint Details

### Route

```
POST /api/v1/accounts/auth/refresh
```

### Request

```json
{
  "refresh-token": "<opaque-token>"
}
```

**Validation:**

- `refresh-token`: Required, must be a valid base64url-encoded string

### Response (Success - 200 OK)

```json
{
  "access-token": "<new-jwt-token>",
  "refresh-token": "<new-opaque-token>"
}
```

### Error Responses

| Status                    | Condition                                           |
| ------------------------- | --------------------------------------------------- |
| 400 Bad Request           | Invalid JSON, missing field, or validation failure  |
| 401 Unauthorized          | Token not found, expired, revoked, or user inactive |
| 500 Internal Server Error | Database or service error                           |

## Security Implementation

### Token Generation

- **Algorithm**: Cryptographically secure random generation using `crypto/rand`
- **Size**: 48 bytes for sufficient entropy
- **Encoding**: Base64 URL-safe encoding for safe transmission
- **Format**: Opaque token (no structured data exposed)

### Token Storage

- **Mechanism**: SHA256 hashing before database storage
- **Lookup**: Hash-based query prevents reverse lookups
- **Plaintext**: Never stored in database (only hash)

### Token Lifecycle Management

1. **Old Token Revocation**: Previous session revoked immediately after new token generation
2. **Atomic Operation**: Revocation and new session creation wrapped in transaction
3. **Expiration Tracking**: Session expiry set to 7 days from creation
4. **Status Validation**: User must be active to refresh

### Validation Chain

```
Request → Format Validation
   ↓
Empty Check (raise ErrInvalidToken)
   ↓
SHA256 Hash
   ↓
Session Lookup (ErrInvalidToken if not found)
   ↓
Session Validity Check (revoked_at IS NULL AND expires_at > NOW())
   ↓
User Existence Check (raise ErrInvalidToken if not found)
   ↓
User Active Status Check (raise ErrInvalidToken if inactive)
   ↓
Generate New Tokens (access + refresh)
   ↓
Atomic Transaction:
   ├─ Revoke Old Session
   └─ Create New Session with Hashed Token
   ↓
Return New Tokens
```

## Implementation Flow

### Request Processing

1. **Controller** (`RefreshByToken` handler):
   - Reads JSON request
   - Validates request structure (base64rawurl format)
   - Calls `AuthService.RefreshAuthByToken()`
   - Returns formatted JSON response with new tokens

2. **Service** (`AuthService.RefreshAuthByToken()`):
   - Validates token not empty
   - Hashes token with SHA256
   - Retrieves session from database
   - Validates user active status
   - Generates new secure tokens
   - Executes transaction to revoke old + create new session
   - Returns both tokens

3. **Repository** (`SessionRepository`):
   - `GetByRefreshToken()`: Queries by hashed token, validates not revoked and not expired
   - `Revoke()`: Marks session with revoked_at timestamp
   - `Create()`: Inserts new session with hashed token and expiry

## Code Changes

### File: [internal/controller/auth.go](../internal/controller/auth.go)

#### 1. New Request Struct

```go
type refreshRequest struct {
    RefreshToken string `json:"refresh_token" validate:"required,base64rawurl"`
}
```

#### 2. New Handler Method

```go
func (c *AccountController) RefreshByToken(w http.ResponseWriter, r *http.Request) {
    var req refreshRequest
    if err := jsonutil.ReadJSON(w, r, &req); err != nil {
        jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
        return
    }

    if err := validateStruct(req); err != nil {
        jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
        return
    }

    result, err := c.authSvc.RefreshAuthByToken(r.Context(), req.RefreshToken)
    if err != nil {
        jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
        return
    }

    envelope := struct {
        AccessToken  string `json:"access_token"`
        RefreshToken string `json:"refresh_token"`
    }{
        AccessToken:  result.AccessToken,
        RefreshToken: result.RefreshToken,
    }
    if err = jsonutil.WriteJSON(w, http.StatusOK, envelope); err != nil {
        jsonutil.WriteErrJSON(w, http.StatusInternalServerError,
            domain.NewError(domain.ErrorCode("WRITE_RESPONSE_FAILED"),
            "failed to write response"))
    }
}
```

#### 3. New Route Registration

```go
func SetupAccountRoutes(r chi.Router, authService *auth.AuthService) {
    accountController := NewAccountController(authService)

    r.Route("/auth", func(r chi.Router) {
            r.Route("/by-otp", func(r chi.Router) {
                r.Post("/request", accountController.RequestAuthByOTP)
                r.Post("/verify", accountController.VerifyAuthByOTP)
            })
            r.Post("/refresh", accountController.RefreshByToken)  // ← NEW
        })
}
```

## Architecture & Design Patterns

### Layers Involved

1. **HTTP Controller Layer**: Request/response handling and validation
2. **Service Layer**: Business logic and orchestration
3. **Repository Layer**: Data access and persistence
4. **Domain Layer**: Models and interfaces

### Design Principles Applied

- **Single Responsibility**: Each component has one clear purpose
- **Dependency Injection**: All dependencies passed to service constructor
- **Transaction Safety**: Atomic revoke + create operation
- **Error Handling**: Proper error mapping to HTTP status codes
- **Input Validation**: Multi-level validation (struct tags + business logic)
- **Security by Default**: Hash-based storage, secure token generation

## Token Revocation Strategy

The implementation uses a **timestamp-based revocation** strategy:

- **Approach**: Sets `revoked_at` timestamp when revoking
- **Query Impact**: Lookup queries check `revoked_at IS NULL`
- **Benefits**:
  - Audit trail (can see when token was revoked)
  - No data deletion (complies with retention policies)
  - Atomic with new session creation

## Error Handling Map

| Error             | HTTP Status               | Meaning                                                         |
| ----------------- | ------------------------- | --------------------------------------------------------------- |
| `ErrInvalidToken` | 401 Unauthorized          | Token invalid, expired, revoked, or user doesn't exist/inactive |
| `ErrNotFound`     | 401 Unauthorized          | Session not found (mapped as ErrInvalidToken)                   |
| `ErrInternal`     | 500 Internal Server Error | Database or system error                                        |

## Existing Service Layer

The service implementation (`AuthService.RefreshAuthByToken()`) was already correctly implemented with:

- ✅ Secure token generation
- ✅ SHA256 hashing
- ✅ Proper revocation logic
- ✅ Transaction management
- ✅ User status validation
- ✅ Complete error handling

**No changes were needed to the service layer** - only the controller and routing were added.

## Testing Recommendations

### Happy Path

```bash
1. Login to get initial refresh-token
2. Call POST /api/v1/auth/refresh with refresh-token
3. Verify response contains new access_token and refresh_token
4. Verify old refresh-token no longer works
```

### Edge Cases

- Empty refresh-token → 400 Bad Request
- Invalid/malformed token → 401 Unauthorized
- Expired token → 401 Unauthorized
- Already revoked token → 401 Unauthorized
- User deleted between generation and refresh → 401 Unauthorized
- User deactivated between generation and refresh → 401 Unauthorized

## Performance Considerations

- **Database Query**: Single indexed query by `refresh_token_hash`
- **Transaction**: Two operations (revoke + create) in single transaction
- **Token Generation**: Async-safe, minimal overhead
- **Hash Computation**: SHA256 standard, negligible performance impact

## Compliance & Standards

- **JWT Access Token**: Standard JWT format with claims validation
- **Refresh Token**: Opaque token (no claims, harder to forge)
- **OWASP**: Follows token-based authentication best practices
- **RFC 6234**: SHA256 for secure hashing
- **RFC 4648**: Base64url encoding

## Future Enhancements

1. **Token Rotation Strategy**: Implement more aggressive rotation
2. **Revocation List**: Add Redis-backed blacklist for immediate revocation
3. **Rate Limiting**: Prevent token refresh abuse
4. **Device Tracking**: Associate tokens with device fingerprints
5. **Token Expiration Alerts**: Notify users of expiring tokens
