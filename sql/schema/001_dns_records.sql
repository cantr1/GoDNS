-- +goose Up
CREATE TABLE dns_records (
    id UUID primary key,
    name TEXT NOT NULL,
    ttl INTEGER NOT NULL ,
    class TEXT NOT NULL,
    type TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE dns_records;