package service

import (
	"encoding/json"
	"net/http"

	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/openclick/internal/models/filter"
)

// ─────────────────────────────────────────────────────────────────────────────
// Analytics Query Handlers
// ─────────────────────────────────────────────────────────────────────────────

// QueryTrends handles POST /api/v1/projects/:project_id/query/trends
func (s *Service) QueryTrends(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	if !s.checkAnalyticsAuth(w, r, projectID, "analytics:read") {
		return
	}

	if s.analyticsDB == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"results": []interface{}{}})
		return
	}

	var body struct {
		Events    []filter.TrendsEvent   `json:"events"`
		DateFrom  string                 `json:"date_from"`
		DateTo    string                 `json:"date_to"`
		Interval  string                 `json:"interval"`
		Filters   []filter.PropertyFilter `json:"filters"`
		Breakdown string                 `json:"breakdown"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	result, err := s.analyticsDB.QueryTrends(ctx, &filter.TrendsQuery{
		ProjectID: projectID,
		Events:    body.Events,
		DateFrom:  body.DateFrom,
		DateTo:    body.DateTo,
		Interval:  body.Interval,
		Filters:   body.Filters,
		Breakdown: body.Breakdown,
	})
	if err != nil {
		logger.Error(ctx, "query trends: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to query trends", "INTERNAL_ERROR")
		return
	}

	// Build response
	var results []map[string]interface{}
	for _, series := range result.Results {
		results = append(results, map[string]interface{}{
			"label":           series.Label,
			"breakdown_value": series.BreakdownValue,
			"data":            series.Data,
			"labels":          series.Labels,
			"days":            series.Days,
		})
	}
	if results == nil {
		results = []map[string]interface{}{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"results": results})
}

// QueryFunnel handles POST /api/v1/projects/:project_id/query/funnel
func (s *Service) QueryFunnel(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	if !s.checkAnalyticsAuth(w, r, projectID, "analytics:read") {
		return
	}

	if s.analyticsDB == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"result": []interface{}{}})
		return
	}

	var body struct {
		Steps                []filter.FunnelStep    `json:"steps"`
		DateFrom             string                 `json:"date_from"`
		DateTo               string                 `json:"date_to"`
		ConversionWindowDays int                    `json:"conversion_window_days"`
		FunnelOrder          string                 `json:"funnel_order"`
		Filters              []filter.PropertyFilter `json:"filters"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	result, err := s.analyticsDB.QueryFunnel(ctx, &filter.FunnelQuery{
		ProjectID:            projectID,
		Steps:                body.Steps,
		DateFrom:             body.DateFrom,
		DateTo:               body.DateTo,
		ConversionWindowDays: body.ConversionWindowDays,
		FunnelOrder:          body.FunnelOrder,
		Filters:              body.Filters,
	})
	if err != nil {
		logger.Error(ctx, "query funnel: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to query funnel", "INTERNAL_ERROR")
		return
	}

	var steps []map[string]interface{}
	for _, step := range result.Result {
		steps = append(steps, map[string]interface{}{
			"action_id":               step.ActionID,
			"name":                    step.Name,
			"count":                   step.Count,
			"conversion_rate":         step.ConversionRate,
			"average_conversion_time": step.AverageConversionTime,
		})
	}
	if steps == nil {
		steps = []map[string]interface{}{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"result": steps})
}

// QueryRetention handles POST /api/v1/projects/:project_id/query/retention
func (s *Service) QueryRetention(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	if !s.checkAnalyticsAuth(w, r, projectID, "analytics:read") {
		return
	}

	if s.analyticsDB == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"result": []interface{}{}})
		return
	}

	var body struct {
		TargetEvent   filter.RetentionEvent `json:"target_event"`
		ReturnEvent   filter.RetentionEvent `json:"return_event"`
		DateFrom      string                `json:"date_from"`
		DateTo        string                `json:"date_to"`
		Period        string                `json:"period"`
		RetentionType string                `json:"retention_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	result, err := s.analyticsDB.QueryRetention(ctx, &filter.RetentionQuery{
		ProjectID:     projectID,
		TargetEvent:   body.TargetEvent,
		ReturnEvent:   body.ReturnEvent,
		DateFrom:      body.DateFrom,
		DateTo:        body.DateTo,
		Period:        body.Period,
		RetentionType: body.RetentionType,
	})
	if err != nil {
		logger.Error(ctx, "query retention: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to query retention", "INTERNAL_ERROR")
		return
	}

	var rows []map[string]interface{}
	for _, cohort := range result.Result {
		var values []map[string]interface{}
		for _, v := range cohort.Values {
			values = append(values, map[string]interface{}{
				"count":      v.Count,
				"percentage": v.Percentage,
			})
		}
		rows = append(rows, map[string]interface{}{
			"date":        cohort.Date,
			"label":       cohort.Label,
			"cohort_size": cohort.CohortSize,
			"values":      values,
		})
	}
	if rows == nil {
		rows = []map[string]interface{}{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"result": rows})
}

// QueryPaths handles POST /api/v1/projects/:project_id/query/paths
func (s *Service) QueryPaths(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	if !s.checkAnalyticsAuth(w, r, projectID, "analytics:read") {
		return
	}

	if s.analyticsDB == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"nodes": []interface{}{}, "links": []interface{}{}})
		return
	}

	var body struct {
		DateFrom      string `json:"date_from"`
		DateTo        string `json:"date_to"`
		StartPoint    string `json:"start_point"`
		EndPoint      string `json:"end_point"`
		PathType      string `json:"path_type"`
		StepLimit     int    `json:"step_limit"`
		MinEdgeWeight int    `json:"min_edge_weight"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	result, err := s.analyticsDB.QueryPaths(ctx, &filter.PathsQuery{
		ProjectID:     projectID,
		DateFrom:      body.DateFrom,
		DateTo:        body.DateTo,
		StartPoint:    body.StartPoint,
		EndPoint:      body.EndPoint,
		PathType:      body.PathType,
		StepLimit:     body.StepLimit,
		MinEdgeWeight: body.MinEdgeWeight,
	})
	if err != nil {
		logger.Error(ctx, "query paths: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to query paths", "INTERNAL_ERROR")
		return
	}

	var nodes, links []map[string]interface{}
	for _, n := range result.Nodes {
		nodes = append(nodes, map[string]interface{}{"id": n.ID, "name": n.Name})
	}
	for _, l := range result.Links {
		links = append(links, map[string]interface{}{"source": l.Source, "target": l.Target, "value": l.Value})
	}
	if nodes == nil {
		nodes = []map[string]interface{}{}
	}
	if links == nil {
		links = []map[string]interface{}{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"nodes": nodes, "links": links})
}

// QueryEvents handles POST /api/v1/projects/:project_id/query/events
func (s *Service) QueryEvents(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	if !s.checkAnalyticsAuth(w, r, projectID, "events:read") {
		return
	}

	if s.analyticsDB == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"results": []interface{}{}, "total": 0})
		return
	}

	var body struct {
		Event      string                 `json:"event"`
		DateFrom   string                 `json:"date_from"`
		DateTo     string                 `json:"date_to"`
		DistinctID *string                `json:"distinct_id"`
		Filters    []filter.PropertyFilter `json:"filters"`
		Limit      int                    `json:"limit"`
		Offset     int                    `json:"offset"`
		OrderBy    string                 `json:"order_by"`
		OrderDir   string                 `json:"order_dir"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	distinctID := ""
	if body.DistinctID != nil {
		distinctID = *body.DistinctID
	}

	result, err := s.analyticsDB.QueryEvents(ctx, &filter.EventsQuery{
		ProjectID:  projectID,
		Event:      body.Event,
		DateFrom:   body.DateFrom,
		DateTo:     body.DateTo,
		DistinctID: distinctID,
		Filters:    body.Filters,
		Limit:      body.Limit,
		Offset:     body.Offset,
		OrderBy:    body.OrderBy,
		OrderDir:   body.OrderDir,
	})
	if err != nil {
		logger.Error(ctx, "query events: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to query events", "INTERNAL_ERROR")
		return
	}

	var events []map[string]interface{}
	for _, e := range result.Results {
		events = append(events, map[string]interface{}{
			"uuid":        e.UUID,
			"event":       e.Event,
			"distinct_id": e.DistinctID,
			"timestamp":   e.Timestamp,
			"properties":  json.RawMessage(e.Properties),
		})
	}
	if events == nil {
		events = []map[string]interface{}{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": events,
		"total":   result.Total,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Sessions
// ─────────────────────────────────────────────────────────────────────────────

// ListSessions handles GET /api/v1/projects/:project_id/sessions
func (s *Service) ListSessions(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	if !s.checkAnalyticsAuth(w, r, projectID, "replay:read") {
		return
	}

	if s.analyticsDB == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"results": []interface{}{}, "total": 0})
		return
	}

	q := r.URL.Query()
	f := &filter.SessionFilter{
		ProjectID:     projectID,
		DateFrom:      q.Get("date_from"),
		DateTo:        q.Get("date_to"),
		DistinctID:    q.Get("distinct_id"),
		MinDurationMs: queryInt(q, "min_duration_ms", 0),
		Search:        q.Get("search"),
		Limit:         queryInt(q, "limit", 50),
		Offset:        queryInt(q, "offset", 0),
	}

	sessions, total, err := s.analyticsDB.ListSessions(ctx, f)
	if err != nil {
		logger.Error(ctx, "list sessions: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list sessions", "INTERNAL_ERROR")
		return
	}

	var results []map[string]interface{}
	for _, sess := range sessions {
		results = append(results, daoSessionToResponse(sess))
	}
	if results == nil {
		results = []map[string]interface{}{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"results": results, "total": total})
}

// GetSession handles GET /api/v1/projects/:project_id/sessions/:session_id
func (s *Service) GetSession(w http.ResponseWriter, r *http.Request, projectID, sessionID string) {
	ctx := r.Context()
	if !s.checkAnalyticsAuth(w, r, projectID, "replay:read") {
		return
	}

	if s.analyticsDB == nil {
		writeError(w, http.StatusServiceUnavailable, "analytics database not configured", "SERVICE_UNAVAILABLE")
		return
	}

	sess, err := s.analyticsDB.GetSession(ctx, projectID, sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}
	writeJSON(w, http.StatusOK, daoSessionToResponse(sess))
}

// GetSessionChunks handles GET /api/v1/projects/:project_id/sessions/:session_id/chunks
func (s *Service) GetSessionChunks(w http.ResponseWriter, r *http.Request, projectID, sessionID string) {
	ctx := r.Context()
	if !s.checkAnalyticsAuth(w, r, projectID, "replay:read") {
		return
	}

	if s.analyticsDB == nil {
		writeError(w, http.StatusServiceUnavailable, "analytics database not configured", "SERVICE_UNAVAILABLE")
		return
	}

	fromChunk := queryInt(r.URL.Query(), "from_chunk", 0)
	chunks, total, err := s.analyticsDB.GetSessionChunks(ctx, projectID, sessionID, fromChunk)
	if err != nil {
		logger.Error(ctx, "get session chunks: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to get session chunks", "INTERNAL_ERROR")
		return
	}

	var chunkList []map[string]interface{}
	for _, c := range chunks {
		chunkList = append(chunkList, map[string]interface{}{
			"chunk_index": c.ChunkIndex,
			"data":        c.Data,
			"timestamp":   c.Timestamp,
		})
	}
	if chunkList == nil {
		chunkList = []map[string]interface{}{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"chunks":       chunkList,
		"total_chunks": total,
	})
}

// DeleteSession handles DELETE /api/v1/projects/:project_id/sessions/:session_id
func (s *Service) DeleteSession(w http.ResponseWriter, r *http.Request, projectID, sessionID string) {
	ctx := r.Context()
	if !s.checkAnalyticsAuth(w, r, projectID, "replay:delete") {
		return
	}

	if s.analyticsDB == nil {
		writeError(w, http.StatusServiceUnavailable, "analytics database not configured", "SERVICE_UNAVAILABLE")
		return
	}

	if err := s.analyticsDB.DeleteSession(ctx, projectID, sessionID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// checkAnalyticsAuth validates user membership + permission for analytics endpoints
func (s *Service) checkAnalyticsAuth(w http.ResponseWriter, r *http.Request, projectID, perm string) bool {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return false
	}
	if !hasPermission(r, perm) {
		writeError(w, http.StatusForbidden, "missing permission: "+perm, "FORBIDDEN")
		return false
	}
	return s.assertMembership(ctx, w, projectID, userID)
}
