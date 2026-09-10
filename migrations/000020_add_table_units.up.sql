CREATE TABLE units (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    building_id  BIGINT NOT NULL REFERENCES buildings (id),
    floor_id     BIGINT NOT NULL REFERENCES floors (id),
    unit_number  TEXT NOT NULL,
    caption      TEXT,
    status       VARCHAR(20) NOT NULL DEFAULT 'vacant',
    has_parking  BOOLEAN NOT NULL DEFAULT false,
    version      INT NOT NULL DEFAULT 1,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,
    CONSTRAINT chk_unit_number_length CHECK (LENGTH(unit_number) BETWEEN 1 AND 20),
    CONSTRAINT chk_unit_caption_length CHECK (caption IS NULL OR LENGTH(caption) BETWEEN 1 AND 100),
    CONSTRAINT chk_unit_status CHECK (status IN ('vacant', 'occupied'))
);

CREATE UNIQUE INDEX uq_unit_floor_number_active
    ON units (floor_id, unit_number) WHERE deleted_at IS NULL;

CREATE INDEX idx_unit_building_active       
    ON units (building_id) WHERE deleted_at IS NULL;

-- units.building_id is denormalized (read on every unit-level authorization
-- check, avoiding a floor join there) but must never drift from
-- floors.building_id -- enforce it in the DB rather than trusting every
-- caller to set it correctly.
CREATE OR REPLACE FUNCTION trg_units_set_building_id() RETURNS TRIGGER AS $$
BEGIN
    SELECT building_id INTO NEW.building_id FROM floors WHERE id = NEW.floor_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER units_set_building_id
    BEFORE INSERT OR UPDATE OF floor_id ON units
    FOR EACH ROW EXECUTE FUNCTION trg_units_set_building_id();