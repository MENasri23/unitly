
CREATE TYPE building_membership_status AS ENUM ('active', 'removed');

CREATE TABLE IF NOT EXISTS building_memberships (
    id          BIGSERIAL PRIMARY KEY,
    guid        UUID NOT NULL UNIQUE,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    building_id BIGINT NOT NULL REFERENCES buildings(id) ON DELETE CASCADE,
    role_id     SMALLINT NOT NULL REFERENCES building_roles(id),
    status      building_membership_status NOT NULL,
    invited_by  BIGINT REFERENCES users(id),
    version     INTEGER NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, building_id, role_id)
);

CREATE INDEX idx_memberships_user_active ON building_memberships(user_id) WHERE status = 'active';
CREATE INDEX idx_memberships_bldg_active ON building_memberships(building_id) WHERE status = 'active';