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

// Cohort represents a static or dynamic segment of persons
type Cohort struct {
	ID          string          `db:"id"`
	ProjectID   string          `db:"project_id"`
	Name        string          `db:"name"`
	Filters     json.RawMessage `db:"filters"` // JSONB property filter rules
	PersonCount int             `db:"person_count"`
	CreatedAt   time.Time       `db:"created_at"`
}
