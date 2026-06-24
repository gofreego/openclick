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