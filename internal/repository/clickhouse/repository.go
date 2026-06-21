package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/openclick/internal/models/dao"
	"github.com/gofreego/openclick/internal/models/filter"
	_ "github.com/ClickHouse/clickhouse-go/v2" // ClickHouse database/sql driver registration
)

// Config holds ClickHouse connection configuration
type Config struct {
	DSN          string `yaml:"DSN"`          // e.g. "clickhouse://user:pass@host:9000/dbname"
	MaxOpenConns int    `yaml:"MaxOpenConns"`
	MaxIdleConns int    `yaml:"MaxIdleConns"`
}

// Repository implements the ClickHouse portion of service.Repository
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new ClickHouse repository
func NewRepository(ctx context.Context, cfg *Config) (*Repository, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("ClickHouse DSN is required")
	}
	db, err := sql.Open("clickhouse", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open clickhouse: %w", err)
	}
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping clickhouse: %w", err)
	}
	logger.Info(ctx, "ClickHouse repository connected")
	return &Repository{db: db}, nil
}

// Ping checks the ClickHouse connection
func (r *Repository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// ─────────────────────────────────────────────────────────────────────────────
// Event Ingest
// ─────────────────────────────────────────────────────────────────────────────

// InsertEvents batch-inserts events into ClickHouse
func (r *Repository) InsertEvents(ctx context.Context, events []*dao.Event) error {
	if len(events) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO events (project_id, uuid, event, distinct_id, properties, timestamp, session_id, elements_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("prepare insert events: %w", err)
	}
	defer stmt.Close()

	for _, e := range events {
		_, err := stmt.ExecContext(ctx,
			e.ProjectID, e.UUID, e.Event, e.DistinctID, e.Properties, e.Timestamp,
			e.SessionID, e.ElementsHash,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("exec insert event: %w", err)
		}
	}
	return tx.Commit()
}

// UpsertSession inserts or updates session metadata in ClickHouse
func (r *Repository) UpsertSession(ctx context.Context, s *dao.Session) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (project_id, session_id, distinct_id, start_time, end_time,
		                      duration_ms, page_count, click_count, country_code, browser, os, device_type, recording_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, s.ProjectID, s.SessionID, s.DistinctID, s.StartTime, s.EndTime,
		s.DurationMs, s.PageCount, s.ClickCount, s.CountryCode, s.Browser, s.OS, s.DeviceType, s.RecordingURL,
	)
	return err
}

// InsertReplayChunk inserts a single rrweb replay chunk
func (r *Repository) InsertReplayChunk(ctx context.Context, chunk *dao.ReplayChunk) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO replay_chunks (project_id, session_id, chunk_index, data, compressed, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
	`, chunk.ProjectID, chunk.SessionID, chunk.ChunkIndex, chunk.Data, chunk.Compressed, chunk.Timestamp,
	)
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// Sessions
// ─────────────────────────────────────────────────────────────────────────────

// ListSessions returns paginated sessions for a project
func (r *Repository) ListSessions(ctx context.Context, f *filter.SessionFilter) ([]*dao.Session, int, error) {
	conditions := []string{"project_id = ?"}
	args := []interface{}{f.ProjectID}

	if f.DateFrom != "" {
		conditions = append(conditions, "start_time >= ?")
		args = append(args, f.DateFrom)
	}
	if f.DateTo != "" {
		conditions = append(conditions, "end_time <= ?")
		args = append(args, f.DateTo)
	}
	if f.DistinctID != "" {
		conditions = append(conditions, "distinct_id = ?")
		args = append(args, f.DistinctID)
	}
	if f.MinDurationMs > 0 {
		conditions = append(conditions, "duration_ms >= ?")
		args = append(args, f.MinDurationMs)
	}

	where := strings.Join(conditions, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sessions WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count sessions: %w", err)
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	args = append(args, limit, f.Offset)
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT project_id, session_id, distinct_id, start_time, end_time,
		       duration_ms, page_count, click_count, country_code, browser, os, device_type, recording_url
		FROM sessions WHERE %s
		ORDER BY start_time DESC
		LIMIT ? OFFSET ?
	`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*dao.Session
	for rows.Next() {
		var s dao.Session
		if err := rows.Scan(&s.ProjectID, &s.SessionID, &s.DistinctID, &s.StartTime, &s.EndTime,
			&s.DurationMs, &s.PageCount, &s.ClickCount, &s.CountryCode, &s.Browser, &s.OS, &s.DeviceType, &s.RecordingURL,
		); err != nil {
			logger.Error(ctx, "scan session: %v", err)
			continue
		}
		sessions = append(sessions, &s)
	}
	return sessions, total, rows.Err()
}

// GetSession fetches a single session by session_id
func (r *Repository) GetSession(ctx context.Context, projectID, sessionID string) (*dao.Session, error) {
	var s dao.Session
	err := r.db.QueryRowContext(ctx, `
		SELECT project_id, session_id, distinct_id, start_time, end_time,
		       duration_ms, page_count, click_count, country_code, browser, os, device_type, recording_url
		FROM sessions WHERE project_id = ? AND session_id = ?
	`, projectID, sessionID).Scan(
		&s.ProjectID, &s.SessionID, &s.DistinctID, &s.StartTime, &s.EndTime,
		&s.DurationMs, &s.PageCount, &s.ClickCount, &s.CountryCode, &s.Browser, &s.OS, &s.DeviceType, &s.RecordingURL,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session %s not found", sessionID)
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &s, nil
}

// GetSessionChunks retrieves replay chunks for a session starting from fromChunk index
func (r *Repository) GetSessionChunks(ctx context.Context, projectID, sessionID string, fromChunk int) ([]*dao.ReplayChunk, int, error) {
	// Get total count
	var total int
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM replay_chunks WHERE project_id = ? AND session_id = ?`,
		projectID, sessionID).Scan(&total)

	rows, err := r.db.QueryContext(ctx, `
		SELECT project_id, session_id, chunk_index, data, compressed, timestamp
		FROM replay_chunks
		WHERE project_id = ? AND session_id = ? AND chunk_index >= ?
		ORDER BY chunk_index ASC
	`, projectID, sessionID, fromChunk)
	if err != nil {
		return nil, 0, fmt.Errorf("get session chunks: %w", err)
	}
	defer rows.Close()

	var chunks []*dao.ReplayChunk
	for rows.Next() {
		var c dao.ReplayChunk
		if err := rows.Scan(&c.ProjectID, &c.SessionID, &c.ChunkIndex, &c.Data, &c.Compressed, &c.Timestamp); err != nil {
			logger.Error(ctx, "scan replay chunk: %v", err)
			continue
		}
		chunks = append(chunks, &c)
	}
	return chunks, total, rows.Err()
}

// DeleteSession removes a session and all its replay chunks
func (r *Repository) DeleteSession(ctx context.Context, projectID, sessionID string) error {
	_, err := r.db.ExecContext(ctx, `ALTER TABLE replay_chunks DELETE WHERE project_id = ? AND session_id = ?`, projectID, sessionID)
	if err != nil {
		return fmt.Errorf("delete replay chunks: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `ALTER TABLE sessions DELETE WHERE project_id = ? AND session_id = ?`, projectID, sessionID)
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// Analytics Queries
// ─────────────────────────────────────────────────────────────────────────────

// TrendsResult is the response for a trends query
type TrendsResult struct {
	Results []TrendsSeries
}

// TrendsSeries is a single data series in a trends response
type TrendsSeries struct {
	Label          string
	BreakdownValue string
	Data           []int64
	Labels         []string
	Days           []string
}

// QueryTrends runs a trend query against ClickHouse events
func (r *Repository) QueryTrends(ctx context.Context, q *filter.TrendsQuery) (*TrendsResult, error) {
	result := &TrendsResult{}
	for _, ev := range q.Events {
		var interval string
		switch q.Interval {
		case "hour":
			interval = "toStartOfHour(timestamp)"
		case "week":
			interval = "toMonday(timestamp)"
		case "month":
			interval = "toStartOfMonth(timestamp)"
		default:
			interval = "toDate(timestamp)"
		}

		groupBy := interval
		if q.Breakdown != "" {
			groupBy += ", JSONExtractString(properties, '" + q.Breakdown + "')"
		}

		query := fmt.Sprintf(`
			SELECT %s AS period,
			       %s,
			       count() AS cnt
			FROM events
			WHERE project_id = ?
			  AND event = ?
			  AND timestamp >= ? AND timestamp <= ?
			GROUP BY %s
			ORDER BY period ASC
		`, interval, func() string {
			if q.Breakdown != "" {
				return "JSONExtractString(properties, '" + q.Breakdown + "') AS breakdown_val"
			}
			return "''"
		}(), groupBy)

		args := []interface{}{q.ProjectID, ev.ID, q.DateFrom, q.DateTo}
		rows, err := r.db.QueryContext(ctx, query, args...)
		if err != nil {
			logger.Error(ctx, "trends query for event %s: %v", ev.ID, err)
			continue
		}
		defer rows.Close()

		var series TrendsSeries
		series.Label = ev.Name
		for rows.Next() {
			var period time.Time
			var breakdownVal string
			var cnt int64
			if err := rows.Scan(&period, &breakdownVal, &cnt); err != nil {
				continue
			}
			series.Data = append(series.Data, cnt)
			series.Labels = append(series.Labels, period.Format("2006-01-02"))
			series.Days = append(series.Days, period.Format("2006-01-02"))
			series.BreakdownValue = breakdownVal
		}
		result.Results = append(result.Results, series)
	}
	return result, nil
}

// FunnelResult holds the response for a funnel query
type FunnelResult struct {
	Result []FunnelStep
}

// FunnelStep is a single step result in a funnel response
type FunnelStep struct {
	ActionID              string
	Name                  string
	Count                 int64
	ConversionRate        float64
	AverageConversionTime *float64
}

// QueryFunnel computes funnel conversion rates
func (r *Repository) QueryFunnel(ctx context.Context, q *filter.FunnelQuery) (*FunnelResult, error) {
	result := &FunnelResult{}
	if len(q.Steps) == 0 {
		return result, nil
	}

	// Get count of users who performed each event in order within the date range
	// For simplicity, we count unique users per step (without enforcing strict ordering via window funcs)
	var prevCount int64
	for i, step := range q.Steps {
		var count int64
		err := r.db.QueryRowContext(ctx, `
			SELECT count(DISTINCT distinct_id)
			FROM events
			WHERE project_id = ? AND event = ?
			  AND timestamp >= ? AND timestamp <= ?
		`, q.ProjectID, step.Event, q.DateFrom, q.DateTo).Scan(&count)
		if err != nil {
			logger.Error(ctx, "funnel step %s query: %v", step.Event, err)
			continue
		}

		var convRate float64
		if i == 0 || prevCount == 0 {
			convRate = 100.0
		} else {
			convRate = float64(count) / float64(prevCount) * 100
		}
		prevCount = count

		result.Result = append(result.Result, FunnelStep{
			ActionID:       step.Event,
			Name:           step.Name,
			Count:          count,
			ConversionRate: convRate,
		})
	}
	return result, nil
}

// RetentionResult holds the response for a retention query
type RetentionResult struct {
	Result []RetentionCohort
}

// RetentionCohort is a cohort row in a retention result
type RetentionCohort struct {
	Date        time.Time
	Label       string
	CohortSize  int64
	Values      []RetentionValue
}

// RetentionValue is a single cell in the retention table
type RetentionValue struct {
	Count      int64
	Percentage float64
}

// QueryRetention computes retention cohorts
func (r *Repository) QueryRetention(ctx context.Context, q *filter.RetentionQuery) (*RetentionResult, error) {
	result := &RetentionResult{}

	// Get all distinct_ids that performed target event
	rows, err := r.db.QueryContext(ctx, `
		SELECT distinct_id, min(timestamp) AS first_time
		FROM events
		WHERE project_id = ? AND event = ?
		  AND timestamp >= ? AND timestamp <= ?
		GROUP BY distinct_id
	`, q.ProjectID, q.TargetEvent.ID, q.DateFrom, q.DateTo)
	if err != nil {
		return result, fmt.Errorf("retention target query: %w", err)
	}
	defer rows.Close()

	type cohortUser struct {
		DistinctID string
		FirstTime  time.Time
	}
	var cohortUsers []cohortUser
	for rows.Next() {
		var u cohortUser
		rows.Scan(&u.DistinctID, &u.FirstTime)
		cohortUsers = append(cohortUsers, u)
	}

	// Build a simple single cohort
	cohort := RetentionCohort{
		CohortSize: int64(len(cohortUsers)),
		Label:      "Week 0",
	}
	if len(cohortUsers) > 0 {
		cohort.Date = cohortUsers[0].FirstTime
		cohort.Values = append(cohort.Values, RetentionValue{Count: cohort.CohortSize, Percentage: 100.0})
	}
	result.Result = append(result.Result, cohort)
	return result, nil
}

// PathsResult holds the response for a paths query
type PathsResult struct {
	Nodes []PathNode
	Links []PathLink
}

// PathNode is a node in the paths graph
type PathNode struct {
	ID   string
	Name string
}

// PathLink is a directed edge in the paths graph
type PathLink struct {
	Source string
	Target string
	Value  int64
}

// QueryPaths computes user path analysis
func (r *Repository) QueryPaths(ctx context.Context, q *filter.PathsQuery) (*PathsResult, error) {
	result := &PathsResult{}

	// Query page-view sequences grouped by session
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			JSONExtractString(properties, '$current_url') AS src,
			neighbor(JSONExtractString(properties, '$current_url'), 1) AS tgt,
			count() AS weight
		FROM (
			SELECT distinct_id, session_id, timestamp, properties
			FROM events
			WHERE project_id = ?
			  AND event = 'page_view'
			  AND timestamp >= ? AND timestamp <= ?
			ORDER BY distinct_id, session_id, timestamp
		)
		WHERE tgt != '' AND weight >= ?
		GROUP BY src, tgt
		ORDER BY weight DESC
		LIMIT 50
	`, q.ProjectID, q.DateFrom, q.DateTo, q.MinEdgeWeight)
	if err != nil {
		return result, fmt.Errorf("paths query: %w", err)
	}
	defer rows.Close()

	nodeSet := map[string]bool{}
	for rows.Next() {
		var src, tgt string
		var weight int64
		if err := rows.Scan(&src, &tgt, &weight); err != nil {
			continue
		}
		result.Links = append(result.Links, PathLink{Source: src, Target: tgt, Value: weight})
		nodeSet[src] = true
		nodeSet[tgt] = true
	}
	for node := range nodeSet {
		result.Nodes = append(result.Nodes, PathNode{ID: node, Name: node})
	}
	return result, nil
}

// EventsResult holds the response for a raw events query
type EventsResult struct {
	Results []*dao.Event
	Total   int64
	Next    string
}

// QueryEvents returns raw events with filtering and pagination
func (r *Repository) QueryEvents(ctx context.Context, q *filter.EventsQuery) (*EventsResult, error) {
	conditions := []string{"project_id = ?"}
	args := []interface{}{q.ProjectID}

	if q.Event != "" {
		conditions = append(conditions, "event = ?")
		args = append(args, q.Event)
	}
	if q.DateFrom != "" {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, q.DateFrom)
	}
	if q.DateTo != "" {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, q.DateTo)
	}
	if q.DistinctID != "" {
		conditions = append(conditions, "distinct_id = ?")
		args = append(args, q.DistinctID)
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	r.db.QueryRowContext(ctx, "SELECT count() FROM events WHERE "+where, args...).Scan(&total)

	limit := q.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	orderDir := q.OrderDir
	if orderDir != "asc" {
		orderDir = "desc"
	}
	orderBy := q.OrderBy
	if orderBy == "" {
		orderBy = "timestamp"
	}

	args = append(args, limit, q.Offset)
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT uuid, event, distinct_id, properties, timestamp, session_id, elements_hash
		FROM events WHERE %s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, where, orderBy, orderDir), args...)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	result := &EventsResult{Total: total}
	for rows.Next() {
		var e dao.Event
		e.ProjectID = q.ProjectID
		if err := rows.Scan(&e.UUID, &e.Event, &e.DistinctID, &e.Properties, &e.Timestamp, &e.SessionID, &e.ElementsHash); err != nil {
			logger.Error(ctx, "scan event: %v", err)
			continue
		}
		result.Results = append(result.Results, &e)
	}
	return result, rows.Err()
}
