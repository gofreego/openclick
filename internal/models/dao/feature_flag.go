package dao

import (
	"encoding/json"
	"time"
)

// FeatureFlag represents a feature flag for a project
type FeatureFlag struct {
	ID         string          `db:"id"`
	ProjectID  string          `db:"project_id"`
	Key        string          `db:"key"`         // unique key within project
	Name       string          `db:"name"`        // human-readable name
	Active     bool            `db:"active"`
	RolloutPct int16           `db:"rollout_pct"` // 0-100
	Filters    json.RawMessage `db:"filters"`     // JSONB cohort/property targeting rules
	CreatedAt  time.Time       `db:"created_at"`
}
