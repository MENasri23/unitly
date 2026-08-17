DROP INDEX IF EXISTS idx_sessions_user_active;
 
ALTER TABLE sessions
    DROP COLUMN IF EXISTS ip_address,
    DROP COLUMN IF EXISTS user_agent,
    DROP COLUMN IF EXISTS device_type,
    DROP COLUMN IF EXISTS os_name,
    DROP COLUMN IF EXISTS os_version,
    DROP COLUMN IF EXISTS last_used_at,
    DROP COLUMN IF EXISTS label;
