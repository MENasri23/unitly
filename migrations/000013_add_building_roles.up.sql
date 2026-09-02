CREATE TABLE building_roles (
    id         SMALLINT PRIMARY KEY,
    name       VARCHAR(20) NOT NULL UNIQUE,
    level      SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO building_roles (id, name, level)
VALUES
    (1, 'tenant', 20),
    (2, 'owner', 40),
    (3, 'admin', 100);