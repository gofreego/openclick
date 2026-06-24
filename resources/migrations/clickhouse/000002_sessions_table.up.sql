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
