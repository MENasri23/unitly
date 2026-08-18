CREATE INDEX idx_sessions_expires_at
ON sessions (expires_at);

CREATE INDEX idx_sessions_revoked_at
ON sessions (revoked_at);