package dao

import "time"

// Event represents a raw analytics event (stored in ClickHouse)
type Event struct {
	ProjectID    string    `ch:"project_id"`
	UUID         string    `ch:"uuid"`
	Event        string    `ch:"event"`
	DistinctID   string    `ch:"distinct_id"`
	Properties   string    `ch:"properties"`    // JSON stored as String (device props stripped)
	DeviceID     string    `ch:"device_id"`     // FK to devices table in PostgreSQL
	Timestamp    time.Time `ch:"timestamp"`
	SessionID    string    `ch:"session_id"`    // nullable
	ElementsHash string    `ch:"elements_hash"` // nullable
}

// Session represents session replay metadata (stored in ClickHouse)
type Session struct {
	ProjectID   string    `ch:"project_id"`
	SessionID   string    `ch:"session_id"`
	DistinctID  string    `ch:"distinct_id"`
	StartTime   time.Time `ch:"start_time"`
	EndTime     time.Time `ch:"end_time"`
	DurationMs  uint32    `ch:"duration_ms"`
	PageCount   uint16    `ch:"page_count"`
	ClickCount  uint16    `ch:"click_count"`
	CountryCode string    `ch:"country_code"`
	Browser     string    `ch:"browser"`
	OS          string    `ch:"os"`
	DeviceType  string    `ch:"device_type"`
	RecordingURL string   `ch:"recording_url"`
}

// ReplayChunk is a single rrweb event chunk for session replay (stored in ClickHouse)
type ReplayChunk struct {
	ProjectID  string    `ch:"project_id"`
	SessionID  string    `ch:"session_id"`
	ChunkIndex uint16    `ch:"chunk_index"`
	Data       string    `ch:"data"`       // compressed rrweb JSON chunk
	Compressed bool      `ch:"compressed"`
	Timestamp  time.Time `ch:"timestamp"`
}

// PersonProperty tracks person property history in ClickHouse
type PersonProperty struct {
	ProjectID  string    `ch:"project_id"`
	DistinctID string    `ch:"distinct_id"`
	Key        string    `ch:"key"`
	Value      string    `ch:"value"`
	SetAt      time.Time `ch:"set_at"`
}
