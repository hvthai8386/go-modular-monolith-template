-- Baseline migration. It changes nothing, but gives goose a first version to record so the
-- versioning module can report a schema version from day one. Replace it with real DDL.
-- +goose Up
SELECT 1;

-- +goose Down
SELECT 1;
