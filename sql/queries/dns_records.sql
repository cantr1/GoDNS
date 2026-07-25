-- name: GetDNSRecords :many
SELECT * FROM dns_records;

-- name: GetDNSRecordByName :one
SELECT name, ttl, class, type, value FROM dns_records WHERE name = $1;

-- name: GetDNSRecordByValue :one
SELECT name, ttl, class, type, value  FROM dns_records WHERE value = $1;

-- name: CreateDNSRecord :one
INSERT INTO dns_records (id, name, ttl, class, type, value, created_at, updated_at)
VALUES (
        gen_random_uuid(),$1, $2, $3, $4, $5, NOW(), NOW()
)
RETURNING *;

-- name: RemoveRecords :exec
DELETE FROM dns_records;