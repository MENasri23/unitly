-- Step 1: Drop the index
DROP INDEX IF EXISTS idx_sessions_family_id;

-- Step 2: Drop constraints and columns (PostgreSQL automatically drops constraints associated with the columns)
ALTER TABLE sessions
    DROP COLUMN IF EXISTS parent_session_id,
    DROP COLUMN IF EXISTS family_id;