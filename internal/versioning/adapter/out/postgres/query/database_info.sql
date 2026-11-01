-- name: ServerVersion :one
SELECT version();

-- goose v3 deletes a row on rollback, so the highest version present is the applied schema version.
-- name: SchemaVersion :one
SELECT COALESCE(MAX(version_id), 0)::BIGINT AS version_id FROM goose_db_version;
