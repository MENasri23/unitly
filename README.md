# Unitly

Unitly is a Go-based authentication service focused on OTP-based sign-in and password reset flows.

## Overview

This repository implements:

- OTP request and verification for login/registration
- Password-based authentication for existing active users
- OTP-based password reset flow with secure reset token handling
- JWT access token generation with refresh token support
- Secure refresh token management (opaque tokens with SHA256 hashing)
- Redis-backed OTP session storage and rate limiting
- PostgreSQL user persistence
- Consistent JSON error responses

## Architecture

Key packages:

- `cmd/main.go` - application bootstrap and route registration
- `bootstrap/` - configuration and environment loading
- `internal/controller/` - HTTP controllers and request validation
- `internal/service/` - business logic and services
- `internal/repository/` - database and cache repositories
- `internal/domain/` - domain models, error definitions, and repository interfaces
- `pkg/jsonutil/` - JSON request/response helpers

## Quick Start

### Requirements

- Go 1.25+
- PostgreSQL
- Redis

### Setup

1. Create a `.env` file in the repository root.
2. Populate environment values for PostgreSQL, Redis, JWT secrets, and OTP configuration.
3. Run database migrations from `migrations/`.
4. Start the service using `go run cmd/main.go`.

### Example `.env`

```env
APP_ENV=development
SERVER_HOST=127.0.0.1
SERVER_PORT=8080
WRITE_TIMEOUT=15s
READ_TIMEOUT=15s
READ_HEADER_TIMEOUT=5s
IDLE_TIMEOUT=60s

DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASS=postgres
DB_NAME=unitly

REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_USER=
REDIS_PASS=

ACCESS_TOKEN_SECRET=supersecretaccesskey
ACCESS_TOKEN_EXPIRY=24h

OTP_TEMPLATE_ID=
OTP_RATE_LIMIT=1m

SMS_USER=
SMS_API_KEY=
```

> Note: The SMS provider implementation is currently commented out in `cmd/main.go`. The service uses a mock SMS service for development.

## Server Routes

The HTTP API is mounted at `/api/v1`.

### Auth by OTP

#### Request OTP

- Method: `POST`
- Path: `/api/v1/auth/by-otp/request`
- Request body:
  ```json
  {
    "phone": "+1234567890"
  }
  ```
- Success response: JSON string token
- Use case: creates or refreshes an OTP session for login/registration.

#### Verify OTP

- Method: `POST`
- Path: `/api/v1/auth/by-otp/verify`
- Request body:
  ```json
  {
    "sign": "<request-token>",
    "otp": "123456"
  }
  ```
- Success response:
  ```json
  {
    "access-token": "<jwt-access-token>",
    "refresh-token": "<opaque-refresh-token>"
  }
  ```
- Behavior: if the phone is not found, the service auto-registers a new active user.

### Auth by Password

#### Authenticate with phone and password

- Method: `POST`
- Path: `/api/v1/auth/by-password`
- Request body:
  ```json
  {
    "phone": "+1234567890",
    "password": "StrongP@ssw0rd!"
  }
  ```
- Success response:
  ```json
  {
    "access_token": "<jwt-access-token>",
    "refresh_token": "<opaque-refresh-token>"
  }
  ```
- Behavior:
  - Authenticates only active users who already have a stored password.
  - Returns `401 Unauthorized` with `INVALID_CREDENTIALS` for invalid phone/password combinations, inactive users, or users without a password.
  - This response is intentionally generic to avoid revealing whether a phone number exists in the system.

#### Client integration notes

- Use this endpoint for classic username/password-style sign-in for existing accounts.
- On success, persist both tokens securely on the client and use the access token for authenticated requests.
- On failure, treat the user as unauthenticated and prompt them to retry or start a password reset flow.
- If the client needs to recover access for a user who never set a password, guide them through the OTP-based password reset flow instead.

### Refresh Token

#### Refresh Access Token

- Method: `POST`
- Path: `/api/v1/auth/refresh`
- Request body:
  ```json
  {
    "refresh_token": "<opaque-refresh-token>"
  }
  ```
- Success response:
  ```json
  {
    "access_token": "<new-jwt-access-token>",
    "refresh_token": "<new-opaque-refresh-token>"
  }
  ```
- Behavior:
  - Validates the refresh token and user status
  - Revokes the old refresh token (prevents reuse)
  - Returns new access token and refresh token
  - Fails with `401 Unauthorized` if token is invalid, expired, or user is inactive

### Logout

#### Logout User

- Method: `POST`
- Path: `/api/v1/auth/logout`
- Request body (native clients):
  ```json
  {
    "refresh_token": "<opaque-refresh-token>"
  }
  ```
- Authentication: uses refresh token from request body for native clients or from the `refresh_token` cookie for web clients
- Success response:
  - Web clients: `204 No Content` and auth cookies cleared
  - Native clients: `200 OK` with body `true`
- Behavior:
  - Revokes the refresh token session via the session service
  - Clears auth cookies for web clients
  - Treats already-invalidated refresh tokens as logged out without error

### Password Reset by OTP

#### Request Reset OTP

- Method: `POST`
- Path: `/api/v1/auth/reset-password/by-otp/request`
- Request body:
  ```json
  {
    "phone": "+1234567890"
  }
  ```
- Success response: JSON string token
- Behavior: if the phone is not associated with an active user, the service still returns a token without revealing user existence.

#### Verify Reset OTP

- Method: `POST`
- Path: `/api/v1/auth/reset-password/by-otp/verify`
- Request body:
  ```json
  {
    "sign": "<request-token>",
    "otp": "123456"
  }
  ```
- Success response: JSON string reset token
- Use case: obtain a reset token after OTP verification.

#### Finalize Password Reset

- Method: `POST`
- Path: `/api/v1/auth/reset-password/by-otp/finalize`
- Request body:
  ```json
  {
    "sign": "<reset-token>",
    "password": "NewP@ssw0rd!"
  }
  ```
- Success response: `true`
- Behavior: validates the reset token, updates the user password, and revokes all sessions.

#### Update Full Name

- Method: `PATCH`
- Path: `/api/v1/profiles/change-name`
- Authentication: `Bearer <access-token>`
- Request body:

  ```json
  {
    "first_name": "John",
    "last_name": "Doe"
  }
  ```

  All fields are optional. Any field omitted or set to `null` or an empty string (`""`) is left unchanged.

- Success response:
  - Status: `204 No Content`
  - Response body: _(empty)_

- Validation:
  - At least one of `first_name` or `last_name` must be provided.
  - `first_name` must be 2–20 characters long.
  - `last_name` must be 2–20 characters long.

- Behavior:
  - Updates only the provided fields.
  - The authenticated user is identified from the access token.

#### Get Profile Info

- Method: `GET`
- Path: `/api/v1/profiles/info`
- Authentication: `Bearer <access-token>`
- Success response:
  - Status: `200 OK`
  - Response body:
    ```json
    {
      "first_name": "John",
      "last_name": "Doe",
      "has_password": true,
      "phone": "+1234567890"
    }
    ```

- Behavior:
  - Returns the authenticated user's profile information.
  - `first_name` and `last_name` may be `null` if not set.
  - `has_password` indicates whether a password is set for the account.

#### Update Password

- Method: `PATCH`
- Path: `/api/v1/profiles/change-password`
- Authentication: `Bearer <access-token>`
- Request body:

  ```json
  {
    "old_password": "CurrentP@ssw0rd!",
    "new_password": "NewP@ssw0rd!"
  }
  ```

- Success response:
  - Status: `204 No Content`
  - Response body: _(empty)_

- Validation:
  - `old_password` must be 8–64 characters long.
  - `new_password` must be 8–64 characters long with uppercase, lowercase, digit, and symbol.
  - returns `INCORRECT_PASSWORD` when `old_password` is incorrect.

- Behavior:
  - Updates the authenticated user's password.
  - Requires the current password for verification.
  - The authenticated user is identified from the access token.

#### Update Phone Request

- Method: `POST`
- Path: `/api/v1/profiles/change-phone`
- Authentication: `Bearer <access-token>`
- Request body:

  ```json
  {
    "phone": "+1234567890"
  }
  ```

- Success response:
  - Status: `200 OK`
  - Response body: JSON string token (sign)

- Validation:
  - `phone` must be in valid E.164 format.
  - Phone number must be different from the current one.
  - Phone number must not already be in use by another account.

- Behavior:
  - Initiates a phone number update request.
  - Generates an OTP and sends it to the new phone number.
  - Returns a token (sign) to be used in the verification step.
  - The authenticated user is identified from the access token.

#### Update Phone Verify

- Method: `POST`
- Path: `/api/v1/profiles/change-phone/verify`
- Authentication: `Bearer <access-token>`
- Request body:

  ```json
  {
    "sign": "<token-from-request>",
    "otp": "123456"
  }
  ```

- Success response:
  - Status: `204 No Content`
  - Response body: _(empty)_

- Validation:
  - `sign` must be a valid token from the phone update request.
  - `otp` must be a 6-digit number.
  - OTP must not be expired (3 minute window).

- Behavior:
  - Verifies the OTP sent to the new phone number.
  - Updates the authenticated user's phone number upon successful verification.
  - The authenticated user is identified from the access token.

### Sessions

#### List Active Sessions

- Method: `GET`
- Path: `/api/v1/me/sessions`
- Authentication: `Bearer <access-token>`
- Query parameters:
  - `page` (optional, default: 1) — page number for pagination
  - `page_size` (optional, default: 10, max: 20) — number of sessions per page

- Success response:
  - Status: `200 OK`
  - Response body:
    ```json
    {
      "sessions": [
        {
          "id": "550e8400-e29b-41d4-a716-446655440000",
          "is_current": true,
          "ip_address": "192.168.1.100",
          "device_type": "Desktop",
          "os_name": "Windows",
          "os_version": "11",
          "client_label": "Chrome",
          "created_at": "2024-01-15T10:30:00Z",
          "last_used_at": "2024-01-15T14:22:30Z",
          "expires_at": "2024-01-22T10:30:00Z"
        }
      ],
      "page": 1,
      "page_size": 10,
      "total_count": 5,
      "total_pages": 1
    }
    ```

- Behavior:
  - Returns a paginated list of the authenticated user's active sessions.
  - Sessions are ordered by most recently used first.
  - The `is_current` flag indicates the session of the current request.
  - The authenticated user is identified from the access token.

#### Revoke Session

- Method: `DELETE`
- Path: `/api/v1/me/sessions/{sessionID}`
- Authentication: `Bearer <access-token>`
- Success response:
  - Status: `200 OK` or `204 No Content`
  - Response body: _(empty or success message)_

- Validation:
  - `sessionID` must be a valid UUID.
  - Session must belong to the authenticated user, returns "SESSION_FORBIDDEN" error otherwise.
  - Cannot revoke the current active session(error code: "SESSION_IN_USE").

- Behavior:
  - Revokes a single session by ID (logout from a specific device).
  - The authenticated user is identified from the access token.
  - After revocation, the refresh token associated with the session becomes invalid.


## Error Handling

All errors are returned as JSON objects with a `code` and `message` field.

Example:

```json
{
  "code": "INVALID_PHONE",
  "message": "invalid phone number"
}
```

### Standard error response structure

- `code` — domain error code
- `message` — human-readable error message

### Current mapped error codes

- `INVALID_PHONE` — invalid phone format
- `PHONE_UNCHANGED` — phone number must be different from the current one
- `PHONE_ALREADY_IN_USE` — this phone number is already in use
- `INVALID_PASSWORD` — password failed basic validation rules
- `INCORRECT_PASSWORD` - current password is incorrect
- `INVALID_CREDENTIALS` — phone/password credentials are invalid for this account
- `INVALID_OTP` — OTP is invalid or expired
- `TOO_MANY_ATTEMPTS` — OTP verification rate limit exceeded
- `TOO_MANY_REQUESTS` — OTP request rate limit exceeded
- `INVALID_TOKEN` — invalid or expired JWT / reset token
- `VALIDATION_ERROR` — request validation failed
- `INVALID_JSON` — malformed JSON request
- `EMPTY_FULLNAME` - both first name and last name are empty
- `INTERNAL` — generic server error

### HTTP status codes

- `400 Bad Request` — invalid request body, validation failure, invalid phone, invalid JSON
- `401 Unauthorized` — invalid OTP, invalid token, or invalid phone/password credentials
- `404 Not Found` — user or resource not found
- `429 Too Many Requests` — rate limiting or too many OTP attempts
- `500 Internal Server Error` — unexpected server error

## OTP / Security Behavior

- OTPs are stored in Redis and expire after 3 minutes.
- Failed OTP verification increments an attempt counter.
- After 3 failed OTP attempts, the phone is blocked for 10 minutes.
- Request rate limiting is enforced via a Redis-based rate limiter.
- Password reset flow does not disclose whether a phone exists in the system.

## Token Security

### Refresh Tokens

- Refresh tokens are opaque, cryptographically generated tokens (48 bytes of entropy)
- Tokens are encoded with base64url for safe transmission
- Only the SHA256 hash of the token is persisted in the database (plaintext never stored)
- Refresh tokens expire after 7 days
- When a new token is generated via `/auth/refresh`, the old refresh token is immediately revoked
- Revoked tokens cannot be reused even if not expired
- User must be active to refresh tokens; deleted or deactivated users cannot refresh

### Access Tokens

- Access tokens are JWT tokens with a default expiry of 24 hours
- Access tokens contain the user ID and token type
- Access tokens cannot be revoked (only expired tokens are invalidated)

## Database and Cache

- `internal/repository/user_repo.go` persists users in PostgreSQL.
- `internal/repository/otp/otp_repo.go` persists OTP sessions and rate limits in Redis.
- Migrations are stored in `migrations/`.

## Notes

- The service currently creates a new user automatically during OTP sign-in if the phone does not exist.
- The SMS provider is stubbed in `cmd/main.go` and should be replaced with a real implementation for production.
- JWT secrets and OTP secrets are read from environment variables.

## Development

To run the app locally:

```bash
go run cmd/main.go
```

To build a binary:

```bash
go build -o unitly ./cmd
```

## Dependencies

Primary dependencies in `go.mod`:

- `github.com/go-chi/chi/v5`
- `github.com/go-playground/validator/v10`
- `github.com/golang-jwt/jwt/v5`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/redis/go-redis/v9`
- `github.com/spf13/viper`

---

This README is intended to act as both the project documentation and the reference wiki for current Endpoint behavior, error handling, and configuration.
