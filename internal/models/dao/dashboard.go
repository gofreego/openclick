package dao

import (
	"encoding/json"
	"time"
)

// Dashboard is a named collection of insight panels
type Dashboard struct {
	ID        string          `db:"id"`
	ProjectID string          `db:"project_id"`
	Name      string          `db:"name"`
	Layout    json.RawMessage `db:"layout"` // JSONB grid layout
	CreatedAt time.Time       `db:"created_at"`
}

// DashboardItem is a single insight panel on a dashboard
type DashboardItem struct {
	ID          string          `db:"id"`
	DashboardID string          `db:"dashboard_id"`
	Name        string          `db:"name"`
	Type        string          `db:"type"`     // "trends" | "funnel" | "retention" | "paths" | "sessions"
	Query       json.RawMessage `db:"query"`    // serialised query definition (JSONB)
	Position    json.RawMessage `db:"position"` // {"x":0,"y":0,"w":6,"h":4}
}

// Annotation marks a point in time on analytics charts
type Annotation struct {
	ID         string    `db:"id"`
	ProjectID  string    `db:"project_id"`
	Content    string    `db:"content"`
	Scope      string    `db:"scope"`       // "project"
	DateMarker time.Time `db:"date_marker"`
	CreatedBy  string    `db:"created_by"` // external user ID
	CreatedAt  time.Time `db:"created_at"`
}
