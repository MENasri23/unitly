CREATE TABLE IF NOT EXISTS buildings (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT chk_name_length CHECK (LENGTH(name) <= 100),
    CONSTRAINT chk_address_length CHECK (LENGTH(address) <= 1000)
);