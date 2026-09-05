DROP INDEX IF EXISTS idx_buildings_created_by_name;

CREATE INDEX idx_buildings_created_by 
ON buildings(created_by) 
WHERE deleted_at IS NULL;