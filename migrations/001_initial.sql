-- Reference schema for a PostgreSQL adapter. The default binary uses Memory.
CREATE TABLE IF NOT EXISTS projects (id text PRIMARY KEY, name text NOT NULL, owner text NOT NULL, created_at timestamptz NOT NULL, version bigint NOT NULL DEFAULT 1);
CREATE TABLE IF NOT EXISTS build_definitions (id text PRIMARY KEY, project_id text NOT NULL REFERENCES projects(id), name text NOT NULL, dsl jsonb NOT NULL, created_at timestamptz NOT NULL);
CREATE TABLE IF NOT EXISTS executions (id text PRIMARY KEY, definition_id text NOT NULL REFERENCES build_definitions(id), idempotency_key text UNIQUE, state text NOT NULL, action_key text NOT NULL, result_digest text, created_at timestamptz NOT NULL);
CREATE INDEX IF NOT EXISTS executions_state_idx ON executions(state, created_at);
