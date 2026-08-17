ALTER TABLE sessions
    ADD COLUMN IF NOT EXISTS ip_address   TEXT        NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS user_agent   TEXT        NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS device_type  TEXT        NOT NULL DEFAULT 'Unknown',
    ADD COLUMN IF NOT EXISTS os_name      TEXT        NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS os_version   TEXT        NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS label        TEXT        NOT NULL DEFAULT '';
 
UPDATE sessions
SET last_used_at = created_at WHERE last_used_at IS NULL;

ALTER TABLE sessions
ALTER COLUMN last_used_at SET DEFAULT now();

ALTER TABLE sessions
ALTER COLUMN last_used_at SET NOT NULL;
 
-- Powers "list my active sessions" (WHERE user_id = ? AND revoked_at IS
-- NULL AND expires_at > now() ORDER BY last_used_at DESC).
CREATE INDEX IF NOT EXISTS idx_sessions_user_active
    ON sessions (user_id, last_used_at DESC)
    WHERE revoked_at IS NULL;

