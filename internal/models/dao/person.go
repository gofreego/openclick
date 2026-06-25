package dao

import (
	"encoding/json"
	"time"
)

// Person represents an identified end-user of the tracked application
type Person struct {
	ID         string          `db:"id"`
	ProjectID  string          `db:"project_id"`
	DistinctID string          `db:"distinct_id"`
	Properties json.RawMessage `db:"properties"` // JSONB
	CreatedAt  time.Time       `db:"created_at"`
}

// Device represents a unique client device (browser/app instance) identified
// by a fingerprint or an explicit $device_id from the SDK.
type Device struct {
	ID         string          `db:"id"`
	ProjectID  string          `db:"project_id"`
	Properties json.RawMessage `db:"properties"` // JSONB — browser, OS, screen size, lib, etc.
	CreatedAt  time.Time       `db:"created_at"`
	UpdatedAt  time.Time       `db:"updated_at"`
}

// StatItem is a single value+count pair used for aggregation results.
type StatItem struct {
	Value string
	Count int64
}

// Cohort represents a static or dynamic segment of persons
type Cohort struct {
	ID          string          `db:"id"`
	ProjectID   string          `db:"project_id"`
	Name        string          `db:"name"`
	Filters     json.RawMessage `db:"filters"` // JSONB property filter rules
	PersonCount int             `db:"person_count"`
	CreatedAt   time.Time       `db:"created_at"`
}
