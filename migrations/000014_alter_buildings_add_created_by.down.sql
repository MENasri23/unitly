DROP INDEX IF EXISTS idx_buildings_created_by;

ALTER TABLE buildings
DROP COLUMN created_by;
