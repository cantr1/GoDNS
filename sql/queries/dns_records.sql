-- name: GetDNSRecords :many
SELECT * FROM dns_records;

-- name: GetDNSRecordByName :one
SELECT name, ttl, class, type, value FROM dns_records WHERE name = $1;

-- name: GetDNSRecordByValue :one
SELECT name, ttl, class, type, value  FROM dns_records WHERE value = $1;