CREATE TABLE floors (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    building_id  BIGINT NOT NULL REFERENCES buildings (id),
    floor_number SMALLINT NOT NULL,
    caption      VARCHAR(100),
    version      INT NOT NULL DEFAULT 1,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_floor_building_number_active
    ON floors (building_id, floor_number) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_floor_building_caption_active
ON floors (building_id, caption) WHERE deleted_at IS NULL;

CREATE INDEX idx_floors_building_active ON floors (building_id) WHERE deleted_at IS NULL;