DROP INDEX IF EXISTS idx_buildings_created_by;

CREATE UNIQUE INDEX idx_buildings_created_by_name
ON buildings (created_by, LOWER(name))
WHERE deleted_at IS NULL;