-- goose creates this table itself on first run; it is not a migration. sqlc needs the definition
-- to type-check the schema version query. Keep it in sync with goose's postgres dialect.
CREATE TABLE goose_db_version (
    id          SERIAL PRIMARY KEY,
    version_id  BIGINT NOT NULL,
    is_applied  BOOLEAN NOT NULL,
    tstamp      TIMESTAMP DEFAULT NOW()
);
