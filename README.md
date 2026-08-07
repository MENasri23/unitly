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
- `internal/service/auth/` - business logic for auth flows
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
- Path: `/api/v1/accounts/auth/by-otp/request`
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
- Path: `/api/v1/accounts/auth/by-otp/verify`
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
- Path: `/api/v1/accounts/auth/by-password`
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
    "access-token": "<jwt-access-token>",
    "refresh-token": "<opaque-refresh-token>"
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
- Path: `/api/v1/accounts/auth/refresh`
- Request body:
  ```json
  {
    "refresh-token": "<opaque-refresh-token>"
  }
  ```
- Success response:
  ```json
  {
    "access-token": "<new-jwt-access-token>",
    "refresh-token": "<new-opaque-refresh-token>"
  }
  ```
- Behavior: 
  - Validates the refresh token and user status
  - Revokes the old refresh token (prevents reuse)
  - Returns new access token and refresh token
  - Fails with `401 Unauthorized` if token is invalid, expired, or user is inactive

### Password Reset by OTP

#### Request Reset OTP

- Method: `POST`
- Path: `/api/v1/accounts/reset-password/by-otp/request`
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
- Path: `/api/v1/accounts/reset-password/by-otp/verify`
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
- Path: `/api/v1/accounts/reset-password/by-otp/finalize`
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
  - Response body: *(empty)*

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



## Request Validation

The controller validates input using `go-playground/validator` and strict JSON decoding.

- `phone` must conform to E.164 format
- `otp` must be a 6-digit numeric string
- `sign` must be a valid base64url or JWT string depending on the endpoint
- `password` must be 8-64 characters with uppercase, lowercase, digit, and symbol

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
- `INVALID_PASSWORD` — password failed basic validation rules
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