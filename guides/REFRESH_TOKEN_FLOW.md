# Token Refresh Flow Diagram

## Request/Response Flow

```
CLIENT                          CONTROLLER                  SERVICE                 REPOSITORY              DATABASE
  │                                  │                         │                         │                       │
  │ POST /api/v1/accounts/           │                         │                         │                       │
  │ auth/refresh                      │                         │                         │                       │
  │ {                                 │                         │                         │                       │
  │   "refresh-token": "xyz..."       │                         │                         │                       │
  │ }                                 │                         │                         │                       │
  ├─────────────────────────────────>│                         │                         │                       │
  │                                   │ Read & Validate JSON    │                         │                       │
  │                                   │ (base64rawurl check)    │                         │                       │
  │                                   │                         │                         │                       │
  │                                   │ RefreshAuthByToken()    │                         │                       │
  │                                   ├────────────────────────>│                         │                       │
  │                                   │                         │ Validate token ≠ empty │                       │
  │                                   │                         │ SHA256 hash token      │                       │
  │                                   │                         │                         │                       │
  │                                   │                         │ GetByRefreshToken()    │                       │
  │                                   │                         ├────────────────────────>│                       │
  │                                   │                         │                         │ SELECT * FROM        │
  │                                   │                         │                         │   sessions            │
  │                                   │                         │                         │ WHERE hash = ?        │
  │                                   │                         │                         │   AND revoked_at      │
  │                                   │                         │                         │     IS NULL           │
  │                                   │                         │                         │   AND expires_at      │
  │                                   │                         │                         │     > NOW()           │
  │                                   │                         │                         │<────────────────────┤
  │                                   │                         │<────────────────────────┤                       │
  │                                   │                         │ (or ErrInvalidToken)    │                       │
  │                                   │                         │                         │                       │
  │                                   │                         │ GetByGUID(user_id)      │                       │
  │                                   │                         ├────────────────────────>│                       │
  │                                   │                         │                         │ SELECT * FROM users   │
  │                                   │                         │                         │ WHERE id = ?          │
  │                                   │                         │                         │<────────────────────┤
  │                                   │                         │<────────────────────────┤                       │
  │                                   │                         │ Verify user.IsActive    │                       │
  │                                   │                         │                         │                       │
  │                                   │                         │ Generate new tokens:   │                       │
  │                                   │                         │ • New refresh token    │                       │
  │                                   │                         │   (crypto/rand)        │                       │
  │                                   │                         │ • New access token     │                       │
  │                                   │                         │   (JWT)                │                       │
  │                                   │                         │                         │                       │
  │                                   │                         │ BEGIN TRANSACTION       │                       │
  │                                   │                         │ UoW.Do()                │                       │
  │                                   │                         ├────────────────────────>│                       │
  │                                   │                         │ Revoke(old_session_id) │                       │
  │                                   │                         ├────────────────────────>│                       │
  │                                   │                         │                         │ UPDATE sessions       │
  │                                   │                         │                         │ SET revoked_at =      │
  │                                   │                         │                         │     NOW()             │
  │                                   │                         │                         │ WHERE id = ? ...      │
  │                                   │                         │                         │<────────────────────┤
  │                                   │                         │<────────────────────────┤                       │
  │                                   │                         │ Create(new_session)    │                       │
  │                                   │                         │ (with hashed token)    │                       │
  │                                   │                         ├────────────────────────>│                       │
  │                                   │                         │                         │ INSERT INTO sessions  │
  │                                   │                         │                         │ (hash, expiry, ...)  │
  │                                   │                         │                         │<────────────────────┤
  │                                   │                         │ COMMIT TRANSACTION      │                       │
  │                                   │                         │<────────────────────────┤                       │
  │                                   │<────────────────────────┤                         │                       │
  │ {                                 │                         │                         │                       │
  │   "access-token": "eyJ...",       │                         │                         │                       │
  │   "refresh-token": "xyz..."       │ Return JSON response    │                         │                       │
  │ }                                 │ (200 OK)                │                         │                       │
  │<─────────────────────────────────┤                         │                         │                       │
  │                                   │                         │                         │                       │
```

## Security Validation Chain

```
Request Received
    │
    ├─ Format Validation (base64rawurl)
    │   └─ FAIL → 400 Bad Request
    │
    ├─ Empty Check
    │   └─ FAIL → 401 Unauthorized (ErrInvalidToken)
    │
    ├─ SHA256 Hash & Lookup
    │   └─ NOT FOUND → 401 Unauthorized
    │
    ├─ Session Status Check
    │   ├─ Revoked (revoked_at IS NOT NULL)
    │   │   └─ FAIL → 401 Unauthorized
    │   │
    │   └─ Expired (expires_at ≤ NOW())
    │       └─ FAIL → 401 Unauthorized
    │
    ├─ User Exists Check
    │   └─ NOT FOUND → 401 Unauthorized
    │
    └─ User Active Status
        └─ NOT ACTIVE → 401 Unauthorized

    ✓ All checks pass → Generate new tokens
                       Revoke old session (atomic)
                       Return new tokens (200 OK)
```

## Data Flow: Token Lifecycle

```
INITIAL LOGIN/AUTH
│
├─ RequestAuthByOTP → Get OTP token
├─ VerifyAuthByOTP → Verify OTP
└─ Create Session
    ├─ Generate refresh token (opaque, 48 bytes)
    └─ Store SHA256(refresh_token) in DB
        ├─ ExpiresAt = now + 7 days
        └─ CreatedAt = now
            
            ↓

USER CALLS REFRESH ENDPOINT
│
├─ Send plaintext refresh_token
├─ Server SHA256 hashes it
├─ Lookup session by hash
├─ Validate not revoked & not expired
├─ Validate user still active
│
└─ Generate NEW tokens
    ├─ New refresh token (opaque, 48 bytes)
    ├─ New access token (JWT)
    │
    └─ Transaction:
        ├─ Mark old session: revoked_at = now
        ├─ Create new session with SHA256(new_refresh_token)
        │   ├─ ExpiresAt = now + 7 days
        │   └─ CreatedAt = now
        └─ COMMIT
            
            ↓
            
RESPONSE TO CLIENT
│
├─ Return new access_token (JWT)
├─ Return new refresh_token (opaque)
└─ Old refresh_token now invalid

            ↓

OLD TOKEN USED AGAIN?
│
└─ Lookup finds session marked revoked
    └─ Return 401 Unauthorized
```

## Session State Machine

```
                    ┌─────────────────────────┐
                    │  SESSION CREATED        │
                    │  revoked_at: NULL       │
                    │  expires_at: future     │
                    └────────────┬────────────┘
                                 │
                    ┌────────────┴────────────┐
                    │                         │
        ┌───────────▼──────────┐  ┌──────────▼──────────┐
        │ User calls /refresh  │  │ Session expires     │
        │ with refresh_token   │  │ (time passes)       │
        │                      │  │                     │
        └───────────┬──────────┘  └──────────┬──────────┘
                    │                         │
        ┌───────────▼──────────┐  ┌──────────▼──────────┐
        │ NEW TRANSACTION:     │  │ No refresh allowed  │
        │ ├─ Revoke old       │  │ returns 401         │
        │ │  session         │  │                      │
        │ └─ Create new      │  │ Session becomes     │
        │   session         │  │ "expired"            │
        │                  │  │                       │
        └──────────┬──────┘  └──────────┬──────────┘
                   │                    │
        ┌──────────▼──────────┐  ┌──────▼────────────┐
        │ NEW SESSION ACTIVE  │  │ OLD SESSION       │
        │ revoked_at: NULL    │  │ revoked_at: NULL  │
        │ expires_at: future  │  │ expires_at: past  │
        └─────────────────────┘  └───────────────────┘
                   │
                   └─ Cycle continues on refresh
```
