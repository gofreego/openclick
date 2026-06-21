-- OpenClick PostgreSQL Schema Migration
-- Run this against your PostgreSQL database before starting the server.

-- Projects (equivalent to PostHog teams)
CREATE TABLE IF NOT EXISTS projects (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    api_key     TEXT NOT NULL UNIQUE,     -- public key for SDK ingestion: ock_pub_...
    secret_key  TEXT NOT NULL,            -- private key for server-side APIs: ock_sec_...
    timezone    TEXT NOT NULL DEFAULT 'UTC',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Project membership
-- user_id is an opaque external ID from the auth service (no local FK)
CREATE TABLE IF NOT EXISTS project_members (
    project_id  UUID REFERENCES projects(id) ON DELETE CASCADE,
    user_id     TEXT NOT NULL,            -- external user ID from auth service
    role        TEXT NOT NULL DEFAULT 'member',  -- 'owner' | 'member'
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (project_id, user_id)
);

-- Persons (identified end-users of the tracked application)
CREATE TABLE IF NOT EXISTS persons (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    distinct_id TEXT NOT NULL,
    properties  JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, distinct_id)
);

-- Cohorts (static or dynamic user segments)
CREATE TABLE IF NOT EXISTS cohorts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    filters      JSONB NOT NULL DEFAULT '{}',
    person_count INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Feature flags
CREATE TABLE IF NOT EXISTS feature_flags (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id    UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    key           TEXT NOT NULL,
    name          TEXT NOT NULL,
    active        BOOLEAN NOT NULL DEFAULT true,
    rollout_pct   SMALLINT NOT NULL DEFAULT 100 CHECK (rollout_pct BETWEEN 0 AND 100),
    filters       JSONB NOT NULL DEFAULT '{}',   -- cohort/property targeting rules
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, key)
);

-- Dashboards
CREATE TABLE IF NOT EXISTS dashboards (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    layout      JSONB NOT NULL DEFAULT '[]',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Dashboard items (individual insight panels)
CREATE TABLE IF NOT EXISTS dashboard_items (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dashboard_id  UUID NOT NULL REFERENCES dashboards(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    type          TEXT NOT NULL,   -- 'trends' | 'funnel' | 'retention' | 'paths' | 'sessions'
    query         JSONB NOT NULL,  -- serialised query definition
    position      JSONB NOT NULL DEFAULT '{"x":0,"y":0,"w":6,"h":4}'
);

-- Annotations
CREATE TABLE IF NOT EXISTS annotations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    content     TEXT NOT NULL,
    scope       TEXT NOT NULL DEFAULT 'project',
    date_marker TIMESTAMPTZ NOT NULL,
    created_by  TEXT NOT NULL,            -- external user ID from auth service
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_project_members_user_id ON project_members(user_id);
CREATE INDEX IF NOT EXISTS idx_persons_project_id ON persons(project_id);
CREATE INDEX IF NOT EXISTS idx_persons_distinct_id ON persons(project_id, distinct_id);
CREATE INDEX IF NOT EXISTS idx_feature_flags_project_id ON feature_flags(project_id);
CREATE INDEX IF NOT EXISTS idx_dashboards_project_id ON dashboards(project_id);
CREATE INDEX IF NOT EXISTS idx_dashboard_items_dashboard_id ON dashboard_items(dashboard_id);
CREATE INDEX IF NOT EXISTS idx_cohorts_project_id ON cohorts(project_id);
