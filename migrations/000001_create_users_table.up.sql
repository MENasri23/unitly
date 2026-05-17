CREATE TABLE IF NOT EXISTS users(
    id BIGSERIAL PRIMARY KEY,
    public_id UUID NOT NULL UNIQUE,
    phone VARCHAR(20) NOT NULL UNIQUE,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK (password_hash <> '')
);