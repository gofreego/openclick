CREATE TABLE IF NOT EXISTS events (
    project_id String,
    uuid String,
    event String,
    distinct_id String,
    properties String,
    timestamp DateTime,
    session_id String,
    elements_hash String
) ENGINE = MergeTree()
ORDER BY (project_id, event, toDate(timestamp), distinct_id, timestamp);

CREATE TABLE IF NOT EXISTS sessions (
    project_id String,
    session_id String,
    distinct_id String,
    start_time DateTime,
    end_time DateTime,
    duration_ms UInt32,
    page_count UInt16,
    click_count UInt16,
    country_code String,
    browser String,
    os String,
    device_type String,
    recording_url String
) ENGINE = ReplacingMergeTree(end_time)
ORDER BY (project_id, session_id);

CREATE TABLE IF NOT EXISTS replay_chunks (
    project_id String,
    session_id String,
    chunk_index UInt16,
    data String,
    compressed Bool,
    timestamp DateTime
) ENGINE = MergeTree()
ORDER BY (project_id, session_id, chunk_index);

CREATE TABLE IF NOT EXISTS person_properties (
    project_id String,
    distinct_id String,
    key String,
    value String,
    set_at DateTime
) ENGINE = ReplacingMergeTree(set_at)
ORDER BY (project_id, distinct_id, key);
