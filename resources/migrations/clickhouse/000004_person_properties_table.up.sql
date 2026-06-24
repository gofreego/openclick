CREATE TABLE IF NOT EXISTS person_properties (
    project_id String,
    distinct_id String,
    key String,
    value String,
    set_at DateTime
) ENGINE = ReplacingMergeTree(set_at)
ORDER BY (project_id, distinct_id, key);
