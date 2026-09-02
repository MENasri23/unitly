ALTER TABLE buildings
ADD COLUMN created_by BIGINT;

ALTER TABLE buildings
ADD CONSTRAINT fk_buildings_created_by
FOREIGN KEY (created_by)
REFERENCES users(id)
ON DELETE RESTRICT;

CREATE INDEX idx_buildings_created_by ON buildings(created_by) WHERE deleted_at IS NULL;
