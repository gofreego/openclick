# OpenClick — Technical Documentation

> **Version:** 1.0.0-draft  
> **Stack:** Go (backend) · React + TypeScript (frontend)  
> **Status:** Pre-development

---

## Table of Contents

1. [Overview](#overview)
2. [Goals & Non-Goals](#goals--non-goals)
3. [Architecture](#architecture)
4. [Data Models](#data-models)
5. [Authentication & Authorization](#authentication--authorization)
6. [API Definitions](#api-definitions)
   - [Projects](#projects-api)
   - [Events Ingestion](#events-ingestion-api)
   - [Analytics & Funnels](#analytics--funnels-api)
   - [Session Replay](#session-replay-api)
   - [Feature Flags](#feature-flags-api)
   - [Persons & Cohorts](#persons--cohorts-api)
   - [Dashboards](#dashboards-api)
6. [Frontend Architecture](#frontend-architecture)
7. [Infrastructure & Deployment](#infrastructure--deployment)
8. [Performance Targets](#performance-targets)
9. [SDK Specification](#sdk-specification)

---

## Overview

**OpenClick** is an open-source, self-hostable product analytics platform built for speed and low memory footprint. It provides event tracking, session replay, funnel analysis, feature flags, and cohort analytics.

OpenClick is written in **Go** for the backend (single binary, low GC pressure, high concurrency) and **React + TypeScript** for the dashboard UI. The storage layer uses **ClickHouse** for analytics queries and **PostgreSQL** for relational metadata.

---

## Goals & Non-Goals

### Goals

- Drop-in alternative to PostHog for product analytics
- Single Go binary for easy self-hosting (no Kafka, no Redis, no separate Python workers)
- Sub-100ms P99 query latency for most analytics queries
- Memory usage under 256 MB at rest for small/medium deployments
- Full session replay with pixel-perfect rrweb recordings
- Funnel analysis, retention cohorts, and path analysis
- Feature flags with percentage rollouts and user targeting
- JavaScript, Python, Go, and mobile SDKs for event ingestion
- GDPR-compliant by default (data anonymisation, deletion APIs)

### Non-Goals

- Authentication & user management (handled by an external service)
- RBAC / role-based access control (not in v1)
- SAML / SSO (not in v1)
- Multi-tenancy SaaS billing
- Real-time collaborative dashboards
- Data warehouse connectors (Stripe, HubSpot, etc.) in v1

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    OpenClick Binary                     │
│                                                         │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────┐  │
│  │  HTTP/REST  │  │  Ingest      │  │  Batch Worker │  │
│  │  API Server │  │  Pipeline    │  │  (flush loop) │  │
│  │  (Gin)      │  │  (buffered)  │  │               │  │
│  └──────┬──────┘  └──────┬───────┘  └───────┬───────┘  │
│         │                │                  │           │
│         └────────────────┴──────────────────┘           │
│                          │                              │
│              ┌───────────┴────────────┐                 │
│              │    Storage Layer       │                 │
│              │  ┌────────┐ ┌────────┐ │                 │
│              │  │  PG    │ │  CH    │ │                 │
│              │  │(meta)  │ │(events)│ │                 │
│              │  └────────┘ └────────┘ │                 │
│              └────────────────────────┘                 │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│              React + TypeScript Dashboard               │
│         (Vite · TanStack Query · Recharts)              │
└─────────────────────────────────────────────────────────┘
```

### Key Design Decisions

**Single binary.** Unlike PostHog (Python + Node + Kafka + Redis + ClickHouse + Celery), OpenClick ships as one Go binary. Ingest buffering, batch flushing, and background jobs are goroutines — no external queue required for standard deployments.

**ClickHouse for events.** All raw events, session metadata, and replay chunks land in ClickHouse. It handles billions of rows with single-digit millisecond aggregations.

**PostgreSQL for metadata.** Projects, project memberships, feature flags, dashboards, and persons are stored in PostgreSQL. User identity itself lives in an external service — OpenClick only stores external user IDs as references.

**Buffered ingest.** The ingest pipeline writes events to an in-memory ring buffer, flushed to ClickHouse in batches every 1 second or 1,000 events, whichever comes first. This eliminates per-request ClickHouse writes.

---

## Data Models

### PostgreSQL Schemas

```sql
-- Projects (equivalent to PostHog teams)
CREATE TABLE projects (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  api_key     TEXT NOT NULL UNIQUE,     -- public key for SDK ingestion
  secret_key  TEXT NOT NULL,            -- private key for server-side APIs
  timezone    TEXT NOT NULL DEFAULT 'UTC',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Project membership
-- user_id is an opaque external ID from the auth service (no local FK)
CREATE TABLE project_members (
  project_id  UUID REFERENCES projects(id) ON DELETE CASCADE,
  user_id     TEXT NOT NULL,            -- external user ID from auth service
  role        TEXT NOT NULL DEFAULT 'member',  -- 'owner' | 'member'
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, user_id)
);

-- Persons (identified end-users of the tracked application)
CREATE TABLE persons (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  distinct_id TEXT NOT NULL,
  properties  JSONB NOT NULL DEFAULT '{}',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (project_id, distinct_id)
);

-- Feature flags
CREATE TABLE feature_flags (
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
CREATE TABLE dashboards (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  layout      JSONB NOT NULL DEFAULT '[]',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Dashboard items (individual insight panels)
CREATE TABLE dashboard_items (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  dashboard_id  UUID NOT NULL REFERENCES dashboards(id) ON DELETE CASCADE,
  name          TEXT NOT NULL,
  type          TEXT NOT NULL,   -- 'trends' | 'funnel' | 'retention' | 'paths' | 'sessions'
  query         JSONB NOT NULL,  -- serialised query definition
  position      JSONB NOT NULL DEFAULT '{"x":0,"y":0,"w":6,"h":4}'
);

-- Annotations
CREATE TABLE annotations (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  content     TEXT NOT NULL,
  scope       TEXT NOT NULL DEFAULT 'project',
  date_marker TIMESTAMPTZ NOT NULL,
  created_by  TEXT NOT NULL,            -- external user ID from auth service
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### ClickHouse Schemas

```sql
-- Raw events table (primary analytics table)
CREATE TABLE events (
  project_id    String,
  uuid          UUID,
  event         String,
  distinct_id   String,
  properties    String,          -- JSON stored as String for flexibility
  timestamp     DateTime64(3, 'UTC'),
  session_id    Nullable(String),
  elements_hash Nullable(String)
) ENGINE = MergeTree()
PARTITION BY (project_id, toYYYYMM(timestamp))
ORDER BY (project_id, event, toDate(timestamp), distinct_id)
TTL timestamp + INTERVAL 1 YEAR;

-- Persons property history (for cohort queries)
CREATE TABLE person_properties (
  project_id  String,
  distinct_id String,
  key         String,
  value       String,
  set_at      DateTime64(3, 'UTC')
) ENGINE = ReplacingMergeTree(set_at)
ORDER BY (project_id, distinct_id, key);

-- Session replay metadata
CREATE TABLE sessions (
  project_id    String,
  session_id    String,
  distinct_id   String,
  start_time    DateTime64(3, 'UTC'),
  end_time      DateTime64(3, 'UTC'),
  duration_ms   UInt32,
  page_count    UInt16,
  click_count   UInt16,
  country_code  Nullable(String),
  browser       Nullable(String),
  os            Nullable(String),
  device_type   Nullable(String),
  recording_url Nullable(String)  -- object storage path
) ENGINE = ReplacingMergeTree(end_time)
ORDER BY (project_id, session_id);

-- Replay chunks (rrweb events, chunked for streaming)
CREATE TABLE replay_chunks (
  project_id  String,
  session_id  String,
  chunk_index UInt16,
  data        String,           -- compressed rrweb JSON chunk
  compressed  Boolean DEFAULT true,
  timestamp   DateTime64(3, 'UTC')
) ENGINE = MergeTree()
ORDER BY (project_id, session_id, chunk_index)
TTL timestamp + INTERVAL 90 DAY;
```

---

## API Definitions

All API endpoints are prefixed with `/openclick/api/v1`. Ingestion routes (under `/e/`, `/batch/`, `/identify/`, `/alias/`, `/replay/`, `/decide/`) are public-facing and authenticated via the project `api_key`. All other routes are internal/dashboard routes authenticated via headers set by the upstream API gateway.

### Authentication & Authorization

OpenClick does **not** handle user authentication. Identity is resolved upstream (by your API gateway or auth service) and passed to OpenClick via two request headers on every dashboard API call:

| Header | Type | Description |
|---|---|---|
| `x-user-id` | `string` | The authenticated user's ID from the external auth service |
| `x-user-perms` | `string` | Comma-separated permission scopes granted to this user |

OpenClick trusts these headers unconditionally. **Never expose dashboard API routes directly to the internet** — they must sit behind your API gateway which is responsible for verifying identity and setting these headers.

#### Permission Scopes

| Scope | Description |
|---|---|
| `projects:read` | View projects the user is a member of |
| `projects:write` | Create and update projects |
| `projects:delete` | Delete projects |
| `members:write` | Add or remove project members |
| `analytics:read` | Run analytics queries (trends, funnels, retention, paths) |
| `events:read` | Query raw events |
| `replay:read` | View session replays |
| `replay:delete` | Delete session replays |
| `flags:read` | View feature flags |
| `flags:write` | Create and update feature flags |
| `flags:delete` | Delete feature flags |
| `persons:read` | View person profiles |
| `persons:delete` | Delete persons (GDPR erasure) |
| `dashboards:read` | View dashboards |
| `dashboards:write` | Create and edit dashboards |
| `dashboards:delete` | Delete dashboards |

#### Header Example

```http
GET /api/v1/projects HTTP/1.1
x-user-id: usr_abc123
x-user-perms: projects:read,analytics:read,replay:read,flags:read,dashboards:read
```

#### Permission Enforcement Rules

- A request missing `x-user-id` is rejected with `401 Unauthorized`.
- A request to a project endpoint where `x-user-id` is not a member of that project is rejected with `403 Forbidden`, regardless of `x-user-perms`.
- `x-user-perms` is checked on top of membership — a member without `analytics:read` cannot run queries even if they belong to the project.
- The first member added to a project (creator) is automatically assigned the `owner` role. Owners can manage members regardless of `x-user-perms`.

---

Error response format:

```json
{
  "error": "human readable message",
  "code": "MACHINE_READABLE_CODE"
}
```

---

### Projects API

#### `GET /api/v1/projects`

**Required headers:** `x-user-id`, `x-user-perms: projects:read`

Lists all projects where `x-user-id` is a member.

**Response `200`:**
```json
{
  "results": [
    {
      "id": "uuid",
      "name": "My App",
      "api_key": "ock_pub_xxxxx",
      "timezone": "UTC",
      "created_at": "2025-01-01T00:00:00Z"
    }
  ]
}
```

---

#### `POST /api/v1/projects`

Create a new project. The caller (`x-user-id`) is automatically added as the project `owner`.

**Required headers:** `x-user-id`, `x-user-perms: projects:write`

**Request body:**
```json
{
  "name": "My App",
  "timezone": "America/New_York"
}
```

**Response `201`:**
```json
{
  "id": "uuid",
  "name": "My App",
  "api_key": "ock_pub_xxxxx",
  "secret_key": "ock_sec_xxxxx",
  "timezone": "America/New_York",
  "created_at": "2025-01-01T00:00:00Z"
}
```

---

#### `GET /api/v1/projects/:project_id`

Get a single project. Secret key is only returned here.

**Required headers:** `x-user-id`, `x-user-perms: projects:read`

**Response `200`:**
```json
{
  "id": "uuid",
  "name": "My App",
  "api_key": "ock_pub_xxxxx",
  "secret_key": "ock_sec_xxxxx",
  "timezone": "UTC",
  "created_at": "2025-01-01T00:00:00Z",
  "members": [
    { "user_id": "usr_abc123", "role": "owner", "created_at": "2025-01-01T00:00:00Z" }
  ]
}
```

---

#### `PATCH /api/v1/projects/:project_id`

Update project settings.

**Required headers:** `x-user-id`, `x-user-perms: projects:write`

**Request body (all optional):**
```json
{
  "name": "My App v2",
  "timezone": "Europe/London"
}
```

**Response `200`:** Updated project object.

---

#### `DELETE /api/v1/projects/:project_id`

Permanently delete a project and all its data.

**Required headers:** `x-user-id`, `x-user-perms: projects:delete`

**Response `204`:** No content.

---

#### `POST /api/v1/projects/:project_id/members`

Add a member to a project by their external user ID.

**Required headers:** `x-user-id`, `x-user-perms: members:write`

**Request body:**
```json
{
  "user_id": "usr_def456",
  "role": "member"
}
```

**Response `200`:**
```json
{
  "user_id": "usr_def456",
  "role": "member",
  "created_at": "2025-06-01T00:00:00Z"
}
```

---

#### `DELETE /api/v1/projects/:project_id/members/:user_id`

Remove a member from a project. `:user_id` is the external user ID.

**Required headers:** `x-user-id`, `x-user-perms: members:write`

**Response `204`:** No content.

---

### Events Ingestion API

These endpoints are public-facing and authenticated via the project `api_key` passed in the `Authorization: Bearer <api_key>` header or as `api_key` in the request body.

---

#### `POST /e/` — Capture single event

The primary event capture endpoint.

**Request body:**
```json
{
  "api_key": "ock_pub_xxxxx",
  "event": "button_clicked",
  "distinct_id": "user_42",
  "timestamp": "2025-06-01T12:00:00Z",
  "properties": {
    "$current_url": "https://myapp.com/dashboard",
    "$browser": "Chrome",
    "$os": "macOS",
    "plan": "pro",
    "button_label": "Upgrade"
  }
}
```

**Response `200`:**
```json
{ "status": 1 }
```

Notes:
- `timestamp` is optional; defaults to server time.
- `distinct_id` must be non-empty.
- `properties` may contain any JSON-serialisable values.
- Reserved property prefix `$` is used for system properties.

---

#### `POST /batch/` — Capture multiple events

**Request body:**
```json
{
  "api_key": "ock_pub_xxxxx",
  "batch": [
    {
      "event": "page_view",
      "distinct_id": "user_42",
      "timestamp": "2025-06-01T12:00:00Z",
      "properties": { "$current_url": "https://myapp.com/" }
    },
    {
      "event": "signed_up",
      "distinct_id": "user_42",
      "properties": { "plan": "free" }
    }
  ]
}
```

**Response `200`:**
```json
{ "status": 1, "count": 2 }
```

Max batch size: 1,000 events per request.

---

#### `POST /identify/` — Identify a person

Associates a `distinct_id` with person properties.

**Request body:**
```json
{
  "api_key": "ock_pub_xxxxx",
  "distinct_id": "user_42",
  "properties": {
    "$set": {
      "email": "jane@example.com",
      "name": "Jane Smith",
      "plan": "pro"
    },
    "$set_once": {
      "signup_date": "2025-01-01"
    }
  }
}
```

**Response `200`:**
```json
{ "status": 1 }
```

---

#### `POST /alias/` — Alias two identities

Merges an anonymous ID with an identified user.

**Request body:**
```json
{
  "api_key": "ock_pub_xxxxx",
  "alias": "user_42",
  "distinct_id": "anon_abc123"
}
```

**Response `200`:**
```json
{ "status": 1 }
```

---

#### `POST /replay/` — Ingest session replay chunks

Accepts rrweb event chunks for session replay.

**Request body:**
```json
{
  "api_key": "ock_pub_xxxxx",
  "session_id": "sess_xyz",
  "distinct_id": "user_42",
  "chunk_index": 0,
  "data": "<base64-encoded compressed rrweb JSON>",
  "metadata": {
    "start_time": "2025-06-01T12:00:00Z",
    "href": "https://myapp.com/dashboard"
  }
}
```

**Response `200`:**
```json
{ "status": 1 }
```

---

#### `GET /decide/` — Evaluate feature flags for a user

Called client-side to get all active flag values for a given user.

**Query params:**
- `api_key` (required)
- `distinct_id` (required)
- `person_properties` (optional, JSON-encoded)

**Response `200`:**
```json
{
  "feature_flags": {
    "new-checkout-flow": true,
    "dark-mode": false,
    "beta-feature": "variant-b"
  }
}
```

---

### Analytics & Funnels API

All endpoints below require authentication (`Bearer <jwt>`) and are scoped to a project.

---

#### `POST /api/v1/projects/:project_id/query/trends`

Run a trend query (event counts over time).

**Required headers:** `x-user-id`, `x-user-perms: analytics:read`

**Request body:**
```json
{
  "events": [
    {
      "id": "page_view",
      "name": "Page View",
      "math": "total"
    },
    {
      "id": "signed_up",
      "name": "Sign Up",
      "math": "dau"
    }
  ],
  "date_from": "2025-05-01",
  "date_to": "2025-06-01",
  "interval": "day",
  "filters": [
    { "key": "plan", "value": "pro", "operator": "exact", "type": "event" }
  ],
  "breakdown": "$browser"
}
```

`math` options: `total` · `dau` · `wau` · `mau` · `unique_users` · `sum` · `avg` · `min` · `max` · `median` · `p90` · `p95` · `p99`

`interval` options: `hour` · `day` · `week` · `month`

**Response `200`:**
```json
{
  "results": [
    {
      "label": "Page View",
      "breakdown_value": "Chrome",
      "data": [1200, 1350, 980, 1100],
      "labels": ["2025-05-01", "2025-05-02", "2025-05-03", "2025-05-04"],
      "days": ["2025-05-01", "2025-05-02", "2025-05-03", "2025-05-04"]
    }
  ]
}
```

---

#### `POST /api/v1/projects/:project_id/query/funnel`

Run a funnel analysis.

**Required headers:** `x-user-id`, `x-user-perms: analytics:read`

**Request body:**
```json
{
  "steps": [
    { "event": "page_view", "name": "Visited Landing Page" },
    { "event": "signed_up", "name": "Signed Up" },
    { "event": "subscription_started", "name": "Started Subscription" }
  ],
  "date_from": "2025-05-01",
  "date_to": "2025-06-01",
  "conversion_window_days": 14,
  "funnel_order": "ordered",
  "filters": [],
  "breakdown": null
}
```

`funnel_order` options: `ordered` · `unordered` · `strict`

**Response `200`:**
```json
{
  "result": [
    {
      "action_id": "page_view",
      "name": "Visited Landing Page",
      "count": 10000,
      "conversion_rate": 100.0,
      "average_conversion_time": null
    },
    {
      "action_id": "signed_up",
      "name": "Signed Up",
      "count": 3200,
      "conversion_rate": 32.0,
      "average_conversion_time": 45.2
    },
    {
      "action_id": "subscription_started",
      "name": "Started Subscription",
      "count": 800,
      "conversion_rate": 25.0,
      "average_conversion_time": 172.8
    }
  ]
}
```

---

#### `POST /api/v1/projects/:project_id/query/retention`

Compute retention cohorts.

**Required headers:** `x-user-id`, `x-user-perms: analytics:read`

**Request body:**
```json
{
  "target_event": { "id": "signed_up", "name": "Signed Up" },
  "return_event": { "id": "page_view", "name": "Page View" },
  "date_from": "2025-04-01",
  "date_to": "2025-06-01",
  "period": "Week",
  "retention_type": "retention_first_time"
}
```

`period` options: `Day` · `Week` · `Month`

`retention_type` options: `retention_first_time` · `retention_recurring`

**Response `200`:**
```json
{
  "result": [
    {
      "date": "2025-04-01T00:00:00Z",
      "label": "Week 0",
      "cohort_size": 500,
      "values": [
        { "count": 500, "percentage": 100.0 },
        { "count": 220, "percentage": 44.0 },
        { "count": 180, "percentage": 36.0 },
        { "count": 160, "percentage": 32.0 }
      ]
    }
  ]
}
```

---

#### `POST /api/v1/projects/:project_id/query/paths`

User path analysis between events.

**Required headers:** `x-user-id`, `x-user-perms: analytics:read`

**Request body:**
```json
{
  "date_from": "2025-05-01",
  "date_to": "2025-06-01",
  "start_point": "/dashboard",
  "end_point": "/checkout",
  "path_type": "page_view",
  "step_limit": 5,
  "min_edge_weight": 10
}
```

`path_type` options: `page_view` · `custom_event` · `any`

**Response `200`:**
```json
{
  "nodes": [
    { "id": "/dashboard", "name": "/dashboard" },
    { "id": "/settings", "name": "/settings" },
    { "id": "/checkout", "name": "/checkout" }
  ],
  "links": [
    { "source": "/dashboard", "target": "/settings", "value": 320 },
    { "source": "/settings", "target": "/checkout", "value": 180 }
  ]
}
```

---

#### `POST /api/v1/projects/:project_id/query/events`

Query raw events with filters and pagination.

**Required headers:** `x-user-id`, `x-user-perms: events:read`

**Request body:**
```json
{
  "event": "page_view",
  "date_from": "2025-06-01",
  "date_to": "2025-06-02",
  "distinct_id": null,
  "filters": [
    { "key": "$browser", "value": "Chrome", "operator": "exact", "type": "event" }
  ],
  "limit": 100,
  "offset": 0,
  "order_by": "timestamp",
  "order_dir": "desc"
}
```

**Response `200`:**
```json
{
  "results": [
    {
      "uuid": "uuid",
      "event": "page_view",
      "distinct_id": "user_42",
      "timestamp": "2025-06-01T12:00:00Z",
      "properties": {
        "$current_url": "https://myapp.com/dashboard",
        "$browser": "Chrome"
      }
    }
  ],
  "total": 4200,
  "next": "/api/v1/projects/uuid/query/events?offset=100"
}
```

---

### Session Replay API

#### `GET /api/v1/projects/:project_id/sessions`

List recorded sessions with filters.

**Required headers:** `x-user-id`, `x-user-perms: replay:read`

**Query params:**

| Param | Type | Description |
|---|---|---|
| `date_from` | string | ISO 8601 start date |
| `date_to` | string | ISO 8601 end date |
| `distinct_id` | string | Filter by user |
| `min_duration_ms` | int | Minimum session duration |
| `search` | string | Full-text search on URL/properties |
| `limit` | int | Default 50, max 200 |
| `offset` | int | Pagination offset |

**Response `200`:**
```json
{
  "results": [
    {
      "session_id": "sess_xyz",
      "distinct_id": "user_42",
      "start_time": "2025-06-01T12:00:00Z",
      "end_time": "2025-06-01T12:08:45Z",
      "duration_ms": 525000,
      "page_count": 7,
      "click_count": 23,
      "country_code": "IN",
      "browser": "Chrome",
      "os": "macOS",
      "device_type": "desktop"
    }
  ],
  "total": 1823
}
```

---

#### `GET /api/v1/projects/:project_id/sessions/:session_id`

Get session metadata.

**Required headers:** `x-user-id`, `x-user-perms: replay:read`

**Response `200`:** Single session object (same schema as list item).

---

#### `GET /api/v1/projects/:project_id/sessions/:session_id/chunks`

Stream replay chunks for playback.

**Required headers:** `x-user-id`, `x-user-perms: replay:read`

**Query params:**
- `from_chunk` (optional, default 0)

**Response `200`:**
```json
{
  "chunks": [
    {
      "chunk_index": 0,
      "data": "<base64-compressed rrweb JSON>",
      "timestamp": "2025-06-01T12:00:00Z"
    },
    {
      "chunk_index": 1,
      "data": "<base64-compressed rrweb JSON>",
      "timestamp": "2025-06-01T12:00:03Z"
    }
  ],
  "total_chunks": 18
}
```

---

#### `DELETE /api/v1/projects/:project_id/sessions/:session_id`

Delete a session and all its replay chunks.

**Required headers:** `x-user-id`, `x-user-perms: replay:delete`

**Response `204`:** No content.

---

### Feature Flags API

#### `GET /api/v1/projects/:project_id/feature-flags`

List all feature flags.

**Required headers:** `x-user-id`, `x-user-perms: flags:read`

**Response `200`:**
```json
{
  "results": [
    {
      "id": "uuid",
      "key": "new-checkout-flow",
      "name": "New Checkout Flow",
      "active": true,
      "rollout_pct": 50,
      "filters": {
        "groups": [
          {
            "properties": [
              { "key": "plan", "value": "pro", "operator": "exact", "type": "person" }
            ],
            "rollout_percentage": 100
          }
        ]
      },
      "created_at": "2025-05-01T00:00:00Z"
    }
  ]
}
```

---

#### `POST /api/v1/projects/:project_id/feature-flags`

Create a new feature flag.

**Required headers:** `x-user-id`, `x-user-perms: flags:write`

**Request body:**
```json
{
  "key": "new-checkout-flow",
  "name": "New Checkout Flow",
  "active": true,
  "rollout_pct": 50,
  "filters": {
    "groups": [
      {
        "properties": [
          { "key": "plan", "value": ["pro", "enterprise"], "operator": "exact", "type": "person" }
        ],
        "rollout_percentage": 100
      }
    ]
  }
}
```

**Response `201`:** Created flag object.

---

#### `PATCH /api/v1/projects/:project_id/feature-flags/:flag_id`

Update a feature flag.

**Required headers:** `x-user-id`, `x-user-perms: flags:write`

**Request body (all optional):**
```json
{
  "name": "Updated Name",
  "active": false,
  "rollout_pct": 100,
  "filters": {}
}
```

**Response `200`:** Updated flag object.

---

#### `DELETE /api/v1/projects/:project_id/feature-flags/:flag_id`

Delete a feature flag.

**Required headers:** `x-user-id`, `x-user-perms: flags:delete`

**Response `204`:** No content.

---

#### `POST /api/v1/projects/:project_id/feature-flags/evaluate`

Server-side evaluation of flags for a given end-user. Authenticated via `Authorization: Bearer <secret_key>` (project secret key, not a user header — this is called from your backend services).

**Note:** This endpoint uses the project `secret_key` in an `Authorization: Bearer` header instead of `x-user-id`, since it is called server-to-server from your application backend, not from the dashboard.

**Request body:**
```json
{
  "distinct_id": "user_42",
  "person_properties": {
    "plan": "pro",
    "country": "IN"
  },
  "groups": {}
}
```

**Response `200`:**
```json
{
  "feature_flags": {
    "new-checkout-flow": true,
    "dark-mode": false
  }
}
```

---

### Persons & Cohorts API

#### `GET /api/v1/projects/:project_id/persons`

List persons (identified users of the tracked application).

**Required headers:** `x-user-id`, `x-user-perms: persons:read`

**Query params:**

| Param | Type | Description |
|---|---|---|
| `search` | string | Search by distinct_id or email |
| `properties` | string | JSON-encoded property filters |
| `limit` | int | Default 100, max 500 |
| `offset` | int | Pagination offset |

**Response `200`:**
```json
{
  "results": [
    {
      "id": "uuid",
      "distinct_id": "user_42",
      "properties": {
        "email": "jane@example.com",
        "name": "Jane Smith",
        "plan": "pro"
      },
      "created_at": "2025-01-15T00:00:00Z"
    }
  ],
  "total": 12500
}
```

---

#### `GET /api/v1/projects/:project_id/persons/:distinct_id`

Get a single person's profile and recent events.

**Required headers:** `x-user-id`, `x-user-perms: persons:read`

**Response `200`:**
```json
{
  "id": "uuid",
  "distinct_id": "user_42",
  "properties": {
    "email": "jane@example.com",
    "plan": "pro"
  },
  "created_at": "2025-01-15T00:00:00Z",
  "recent_events": [
    {
      "event": "page_view",
      "timestamp": "2025-06-01T12:00:00Z",
      "properties": { "$current_url": "https://myapp.com/dashboard" }
    }
  ]
}
```

---

#### `DELETE /api/v1/projects/:project_id/persons/:distinct_id`

Delete a person and optionally all their events (GDPR right to erasure).

**Required headers:** `x-user-id`, `x-user-perms: persons:delete`

**Query params:**
- `delete_events=true` — also delete all raw events for this user

**Response `204`:** No content.

---

#### `GET /api/v1/projects/:project_id/cohorts`

List defined cohorts.

**Required headers:** `x-user-id`, `x-user-perms: persons:read`

**Response `200`:**
```json
{
  "results": [
    {
      "id": "uuid",
      "name": "Pro Plan Users",
      "filters": {
        "properties": [
          { "key": "plan", "value": "pro", "operator": "exact", "type": "person" }
        ]
      },
      "person_count": 3200,
      "created_at": "2025-03-01T00:00:00Z"
    }
  ]
}
```

---

#### `POST /api/v1/projects/:project_id/cohorts`

Create a static or dynamic cohort.

**Required headers:** `x-user-id`, `x-user-perms: persons:read`

**Request body:**
```json
{
  "name": "Pro Plan Users",
  "filters": {
    "properties": [
      { "key": "plan", "value": "pro", "operator": "exact", "type": "person" }
    ]
  }
}
```

**Response `201`:** Created cohort object.

---

#### `DELETE /api/v1/projects/:project_id/cohorts/:cohort_id`

Delete a cohort.

**Required headers:** `x-user-id`, `x-user-perms: persons:delete`

**Response `204`:** No content.

---

### Dashboards API

#### `GET /api/v1/projects/:project_id/dashboards`

List all dashboards.

**Required headers:** `x-user-id`, `x-user-perms: dashboards:read`

**Response `200`:**
```json
{
  "results": [
    {
      "id": "uuid",
      "name": "Product Overview",
      "item_count": 6,
      "created_at": "2025-01-01T00:00:00Z"
    }
  ]
}
```

---

#### `POST /api/v1/projects/:project_id/dashboards`

Create a dashboard.

**Required headers:** `x-user-id`, `x-user-perms: dashboards:write`

**Request body:**
```json
{
  "name": "Product Overview"
}
```

**Response `201`:** Dashboard object.

---

#### `GET /api/v1/projects/:project_id/dashboards/:dashboard_id`

Get a dashboard with all its items.

**Required headers:** `x-user-id`, `x-user-perms: dashboards:read`

**Response `200`:**
```json
{
  "id": "uuid",
  "name": "Product Overview",
  "items": [
    {
      "id": "uuid",
      "name": "Daily Active Users",
      "type": "trends",
      "query": { "events": [{ "id": "page_view", "math": "dau" }], "date_from": "-30d" },
      "position": { "x": 0, "y": 0, "w": 6, "h": 4 }
    }
  ]
}
```

---

#### `POST /api/v1/projects/:project_id/dashboards/:dashboard_id/items`

Add an insight item to a dashboard.

**Required headers:** `x-user-id`, `x-user-perms: dashboards:write`

**Request body:**
```json
{
  "name": "Daily Active Users",
  "type": "trends",
  "query": {
    "events": [{ "id": "page_view", "math": "dau" }],
    "date_from": "-30d",
    "interval": "day"
  },
  "position": { "x": 0, "y": 0, "w": 6, "h": 4 }
}
```

**Response `201`:** Dashboard item object.

---

#### `PATCH /api/v1/projects/:project_id/dashboards/:dashboard_id/items/:item_id`

Update a dashboard item (query, name, or position).

**Required headers:** `x-user-id`, `x-user-perms: dashboards:write`

**Response `200`:** Updated item object.

---

#### `DELETE /api/v1/projects/:project_id/dashboards/:dashboard_id/items/:item_id`

Remove an item from a dashboard.

**Required headers:** `x-user-id`, `x-user-perms: dashboards:delete`

**Response `204`:** No content.

---

#### `DELETE /api/v1/projects/:project_id/dashboards/:dashboard_id`

Delete a dashboard and all its items.

**Required headers:** `x-user-id`, `x-user-perms: dashboards:delete`

**Response `204`:** No content.

---

## Frontend Architecture

### Tech Stack

| Layer | Choice | Reason |
|---|---|---|
| Bundler | Vite | Fast HMR, ESM native |
| Framework | React 18 | Concurrent features, ecosystem |
| Language | TypeScript 5 | Type safety, DX |
| Routing | React Router v6 | SPA routing |
| State / Cache | TanStack Query v5 | Server state, caching |
| UI Components | shadcn/ui + Radix | Accessible, unstyled primitives |
| Styling | Tailwind CSS | Utility-first, small output |
| Charts | Recharts | React-native, composable |
| Replay Player | rrweb-player | Official rrweb playback |
| Forms | React Hook Form + Zod | Performant, schema validation |
| HTTP Client | Axios | Interceptors, instance config |
| Drag & Drop | dnd-kit | Dashboard layout editing |

### Page Structure

```
src/
├── pages/
│   ├── dashboard/
│   │   ├── DashboardList.tsx
│   │   └── DashboardDetail.tsx
│   ├── events/
│   │   └── EventExplorer.tsx
│   ├── trends/
│   │   └── TrendsInsight.tsx
│   ├── funnels/
│   │   └── FunnelInsight.tsx
│   ├── retention/
│   │   └── RetentionInsight.tsx
│   ├── paths/
│   │   └── PathsInsight.tsx
│   ├── replay/
│   │   ├── SessionList.tsx
│   │   └── SessionPlayer.tsx
│   ├── flags/
│   │   └── FeatureFlags.tsx
│   ├── persons/
│   │   ├── PersonList.tsx
│   │   └── PersonDetail.tsx
│   └── settings/
│       └── ProjectSettings.tsx
├── components/
│   ├── charts/
│   ├── filters/
│   ├── layout/
│   └── shared/
├── hooks/
├── lib/
│   ├── api.ts        # Axios instance + typed API client
│   └── utils.ts
└── types/
    └── index.ts      # Shared TypeScript types
```

---

## Infrastructure & Deployment

### Minimum Self-Hosting Requirements

| Tier | Users | RAM | CPU | Storage |
|---|---|---|---|---|
| Hobby | < 10k MAU | 512 MB | 1 vCPU | 20 GB SSD |
| Small | < 100k MAU | 2 GB | 2 vCPU | 100 GB SSD |
| Medium | < 1M MAU | 8 GB | 4 vCPU | 500 GB SSD |

### Docker Compose (single-node)

```yaml
version: "3.9"
services:
  openclick:
    image: openclick/openclick:latest
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=postgres://oc:oc@postgres:5432/openclick
      - CLICKHOUSE_URL=clickhouse://clickhouse:9000/openclick
      - TRUSTED_PROXIES=172.16.0.0/12
      - OBJECT_STORAGE_URL=file:///data/replays
    volumes:
      - replay_data:/data/replays
    depends_on:
      - postgres
      - clickhouse

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: oc
      POSTGRES_PASSWORD: oc
      POSTGRES_DB: openclick
    volumes:
      - pg_data:/var/lib/postgresql/data

  clickhouse:
    image: clickhouse/clickhouse-server:24
    volumes:
      - ch_data:/var/lib/clickhouse

volumes:
  pg_data:
  ch_data:
  replay_data:
```

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | — | PostgreSQL connection string |
| `CLICKHOUSE_URL` | — | ClickHouse connection string |
| `TRUSTED_PROXIES` | `127.0.0.1` | IPs allowed to set `x-user-id` / `x-user-perms` headers |
| `INGEST_BUFFER_SIZE` | `1000` | Events buffered before flush |
| `INGEST_FLUSH_INTERVAL` | `1s` | Max time between CH flushes |
| `OBJECT_STORAGE_URL` | `file:///data` | Replay storage (`file://`, `s3://`, `gcs://`) |
| `MAX_REPLAY_CHUNK_SIZE` | `512KB` | Max size per replay chunk |
| `SESSION_REPLAY_TTL_DAYS` | `90` | Days to retain replay data |
| `EVENT_TTL_DAYS` | `365` | Days to retain raw events |
| `PORT` | `8000` | HTTP server port |
| `LOG_LEVEL` | `info` | Logging level (`debug`, `info`, `warn`, `error`) |
| `CORS_ORIGINS` | `*` | Allowed CORS origins for ingest |

> **Security note:** Set `TRUSTED_PROXIES` to only the IP(s) of your API gateway. OpenClick will reject requests with `x-user-id` / `x-user-perms` headers that arrive from untrusted IPs, preventing header spoofing from external callers.

---

## Performance Targets

| Metric | Target |
|---|---|
| Event ingest P99 latency | < 5ms (buffered, fire-and-forget) |
| Trend query P99 (30-day range) | < 200ms |
| Funnel query P99 (30-day, 3 steps) | < 500ms |
| Session list query P99 | < 100ms |
| Memory at rest (single binary) | < 256 MB |
| Throughput (single node) | > 10,000 events/sec |
| Replay chunk stream first byte | < 50ms |

---

## SDK Specification

All SDKs share the same behaviour contract.

### JavaScript / TypeScript SDK

```typescript
import OpenClick from '@openclick/js'

const oc = new OpenClick({
  apiKey: 'ock_pub_xxxxx',
  host: 'https://your-openclick-instance.com',
  autocapture: true,          // auto-captures clicks, inputs, page views
  capturePageviews: true,
  sessionRecording: true,
  maskAllInputs: true,        // GDPR: mask all form inputs in replays
  persistence: 'localStorage'
})

// Identify user
oc.identify('user_42', {
  email: 'jane@example.com',
  plan: 'pro'
})

// Track event
oc.capture('button_clicked', {
  button_label: 'Upgrade',
  plan: 'free'
})

// Feature flags
const isEnabled = await oc.isFeatureEnabled('new-checkout-flow')
const variant = await oc.getFeatureFlag('experiment-variant')

// Reset on logout
oc.reset()
```

### Go SDK (server-side)

```go
import "github.com/openclick/openclick-go"

client := openclick.New("ock_sec_xxxxx",
    openclick.WithHost("https://your-openclick.com"),
    openclick.WithFlushInterval(5*time.Second),
)

client.Capture(openclick.Event{
    DistinctID: "user_42",
    Event:      "api_called",
    Properties: map[string]any{
        "endpoint": "/api/orders",
        "method":   "POST",
    },
})

enabled, _ := client.IsFeatureEnabled("new-checkout-flow", "user_42", nil)

client.Flush()
client.Close()
```

### Python SDK (server-side)

```python
import openclick

client = openclick.Client(
    api_key="ock_sec_xxxxx",
    host="https://your-openclick.com"
)

client.capture("user_42", "order_placed", {
    "order_id": "ord_123",
    "total": 99.99,
    "currency": "USD"
})

client.identify("user_42", {
    "$set": {"plan": "pro"}
})

enabled = client.is_feature_enabled("new-checkout-flow", "user_42")

client.flush()
```

---

*OpenClick — built for speed, owned by you.*