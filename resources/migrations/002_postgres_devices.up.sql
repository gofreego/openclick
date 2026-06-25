CREATE TABLE IF NOT EXISTS devices (
    id         TEXT        NOT NULL,
    project_id UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    properties JSONB       NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, id)
);

CREATE INDEX IF NOT EXISTS idx_devices_project_id ON devices(project_id);
