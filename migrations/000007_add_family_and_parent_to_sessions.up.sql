-- Step 1: Add columns
ALTER TABLE sessions
    ADD COLUMN family_id UUID,
    ADD COLUMN parent_session_id UUID;

-- Step 2: Backfill existing rows (if any) so family_id is self-referential for root sessions
UPDATE sessions
SET family_id = id
WHERE family_id IS NULL;

-- Step 3: Enforce NOT NULL on family_id
ALTER TABLE sessions
    ALTER COLUMN family_id SET NOT NULL;

-- Step 4: Add foreign key and unique constraint for parent_session_id
ALTER TABLE sessions
    ADD CONSTRAINT fk_sessions_parent_session_id
        FOREIGN KEY (parent_session_id) 
        REFERENCES sessions(id) 
        ON DELETE CASCADE,
    ADD CONSTRAINT uq_sessions_parent_session_id 
        UNIQUE (parent_session_id);

-- Step 5: Index family_id for fast family lookup and revocation
CREATE INDEX idx_sessions_family_id ON sessions(family_id);