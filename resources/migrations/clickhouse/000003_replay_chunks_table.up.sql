CREATE TABLE IF NOT EXISTS replay_chunks (
    project_id String,
    session_id String,
    chunk_index UInt16,
    data String,
    compressed Bool,
    timestamp DateTime
) ENGINE = MergeTree()
ORDER BY (project_id, session_id, chunk_index);