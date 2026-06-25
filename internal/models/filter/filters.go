package filter

// ProjectFilter filters for listing projects
type ProjectFilter struct {
	UserID string
	Limit  int
	Offset int
}

// PersonFilter filters for listing persons
type PersonFilter struct {
	ProjectID  string
	Search     string // search by distinct_id or email
	Properties string // JSON-encoded property filters
	Limit      int
	Offset     int
}

// SessionFilter filters for listing sessions
type SessionFilter struct {
	ProjectID     string
	DateFrom      string
	DateTo        string
	DistinctID    string
	MinDurationMs int
	Search        string
	Limit         int
	Offset        int
}

// EventsQuery for raw event queries
type EventsQuery struct {
	ProjectID  string
	Event      string
	DateFrom   string
	DateTo     string
	DistinctID string
	Filters    []PropertyFilter
	Limit      int
	Offset     int
	OrderBy    string
	OrderDir   string
}

// PropertyFilter is a single property filter condition
type PropertyFilter struct {
	Key      string      `json:"key"`
	Value    interface{} `json:"value"`
	Operator string      `json:"operator"` // "exact" | "contains" | "gt" | "lt" etc.
	Type     string      `json:"type"`     // "event" | "person"
}

// TrendsQuery for trend analytics
type TrendsQuery struct {
	ProjectID string
	Events    []TrendsEvent
	DateFrom  string
	DateTo    string
	Interval  string // "hour" | "day" | "week" | "month"
	Filters   []PropertyFilter
	Breakdown string
}

// TrendsEvent is a single event series in a trends query
type TrendsEvent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Math string `json:"math"` // "total" | "dau" | "wau" | "mau" etc.
}

// FunnelQuery for funnel analysis
type FunnelQuery struct {
	ProjectID             string
	Steps                 []FunnelStep
	DateFrom              string
	DateTo                string
	ConversionWindowDays  int
	FunnelOrder           string // "ordered" | "unordered" | "strict"
	Filters               []PropertyFilter
	Breakdown             *string
}

// FunnelStep is a step in a funnel
type FunnelStep struct {
	Event string `json:"event"`
	Name  string `json:"name"`
}

// RetentionQuery for retention cohort analysis
type RetentionQuery struct {
	ProjectID     string
	TargetEvent   RetentionEvent
	ReturnEvent   RetentionEvent
	DateFrom      string
	DateTo        string
	Period        string // "Day" | "Week" | "Month"
	RetentionType string // "retention_first_time" | "retention_recurring"
}

// RetentionEvent is an event used in retention analysis
type RetentionEvent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PathsQuery for path analysis
type PathsQuery struct {
	ProjectID     string
	DateFrom      string
	DateTo        string
	StartPoint    string
	EndPoint      string
	PathType      string // "page_view" | "custom_event" | "any"
	StepLimit     int
	MinEdgeWeight int
}

// FeatureFlagFilter for listing feature flags
type FeatureFlagFilter struct {
	ProjectID string
}

// DashboardFilter for listing dashboards
type DashboardFilter struct {
	ProjectID string
}

// CohortFilter for listing cohorts
type CohortFilter struct {
	ProjectID string
}

// DeviceFilter for listing devices
type DeviceFilter struct {
	ProjectID string
	DeviceID  string // exact match; empty means no filter
	Limit     int
	Offset    int
}
