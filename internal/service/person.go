package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/openclick/internal/models/dao"
	"github.com/gofreego/openclick/internal/models/filter"
)

// ─────────────────────────────────────────────────────────────────────────────
// Persons
// ─────────────────────────────────────────────────────────────────────────────

// ListPersons handles GET /api/v1/projects/:project_id/persons
func (s *Service) ListPersons(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "persons:read") {
		writeError(w, http.StatusForbidden, "missing permission: persons:read", "FORBIDDEN")
		return
	}
	if !s.assertMembership(ctx, w, projectID, userID) {
		return
	}

	q := r.URL.Query()
	f := &filter.PersonFilter{
		ProjectID: projectID,
		Search:    q.Get("search"),
		Limit:     queryInt(q, "limit", 100),
		Offset:    queryInt(q, "offset", 0),
	}

	persons, total, err := s.repo.ListPersons(ctx, f)
	if err != nil {
		logger.Error(ctx, "list persons: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list persons", "INTERNAL_ERROR")
		return
	}

	results := make([]map[string]interface{}, 0, len(persons))
	for _, p := range persons {
		results = append(results, personToResponse(p))
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
		"total":   total,
	})
}

// GetPerson handles GET /api/v1/projects/:project_id/persons/:distinct_id
func (s *Service) GetPerson(w http.ResponseWriter, r *http.Request, projectID, distinctID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "persons:read") {
		writeError(w, http.StatusForbidden, "missing permission: persons:read", "FORBIDDEN")
		return
	}
	if !s.assertMembership(ctx, w, projectID, userID) {
		return
	}

	person, err := s.repo.GetPerson(ctx, projectID, distinctID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}

	resp := personToResponse(person)
	// Include recent events if ClickHouse is available
	if s.analyticsDB != nil {
		events, err := s.analyticsDB.QueryEvents(ctx, &filter.EventsQuery{
			ProjectID:  projectID,
			DistinctID: distinctID,
			Limit:      10,
			OrderBy:    "timestamp",
			OrderDir:   "desc",
		})
		if err == nil && events != nil {
			var recentEvents []map[string]interface{}
			for _, e := range events.Results {
				recentEvents = append(recentEvents, map[string]interface{}{
					"event":      e.Event,
					"timestamp":  e.Timestamp,
					"properties": json.RawMessage(e.Properties),
				})
			}
			resp["recent_events"] = recentEvents
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// DeletePerson handles DELETE /api/v1/projects/:project_id/persons/:distinct_id
func (s *Service) DeletePerson(w http.ResponseWriter, r *http.Request, projectID, distinctID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "persons:delete") {
		writeError(w, http.StatusForbidden, "missing permission: persons:delete", "FORBIDDEN")
		return
	}
	if !s.assertMembership(ctx, w, projectID, userID) {
		return
	}

	if err := s.repo.DeletePerson(ctx, projectID, distinctID); err != nil {
		writeError(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─────────────────────────────────────────────────────────────────────────────
// Cohorts
// ─────────────────────────────────────────────────────────────────────────────

// ListCohorts handles GET /api/v1/projects/:project_id/cohorts
func (s *Service) ListCohorts(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "persons:read") {
		writeError(w, http.StatusForbidden, "missing permission: persons:read", "FORBIDDEN")
		return
	}
	if !s.assertMembership(ctx, w, projectID, userID) {
		return
	}

	cohorts, err := s.repo.ListCohorts(ctx, &filter.CohortFilter{ProjectID: projectID})
	if err != nil {
		logger.Error(ctx, "list cohorts: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list cohorts", "INTERNAL_ERROR")
		return
	}

	results := make([]map[string]interface{}, 0, len(cohorts))
	for _, c := range cohorts {
		results = append(results, map[string]interface{}{
			"id":           c.ID,
			"name":         c.Name,
			"filters":      c.Filters,
			"person_count": c.PersonCount,
			"created_at":   c.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"results": results})
}

// CreateCohort handles POST /api/v1/projects/:project_id/cohorts
func (s *Service) CreateCohort(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "persons:read") {
		writeError(w, http.StatusForbidden, "missing permission: persons:read", "FORBIDDEN")
		return
	}
	if !s.assertMembership(ctx, w, projectID, userID) {
		return
	}

	var body struct {
		Name    string          `json:"name"`
		Filters json.RawMessage `json:"filters"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	if body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required", "VALIDATION_ERROR")
		return
	}
	if body.Filters == nil {
		body.Filters = json.RawMessage("{}")
	}

	cohort, err := s.repo.CreateCohort(ctx, &dao.Cohort{
		ProjectID: projectID,
		Name:      body.Name,
		Filters:   body.Filters,
	})
	if err != nil {
		logger.Error(ctx, "create cohort: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create cohort", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":           cohort.ID,
		"name":         cohort.Name,
		"filters":      cohort.Filters,
		"person_count": cohort.PersonCount,
		"created_at":   cohort.CreatedAt,
	})
}

// DeleteCohort handles DELETE /api/v1/projects/:project_id/cohorts/:cohort_id
func (s *Service) DeleteCohort(w http.ResponseWriter, r *http.Request, projectID, cohortID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "persons:delete") {
		writeError(w, http.StatusForbidden, "missing permission: persons:delete", "FORBIDDEN")
		return
	}
	if !s.assertMembership(ctx, w, projectID, userID) {
		return
	}

	if err := s.repo.DeleteCohort(ctx, projectID, cohortID); err != nil {
		writeError(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func personToResponse(p *dao.Person) map[string]interface{} {
	return map[string]interface{}{
		"id":          p.ID,
		"distinct_id": p.DistinctID,
		"properties":  p.Properties,
		"created_at":  p.CreatedAt,
	}
}

// queryInt reads an integer query param with a default value
func queryInt(q interface{ Get(string) string }, key string, def int) int {
	s := q.Get(key)
	if s == "" {
		return def
	}
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return def
	}
	return n
}

// ensure time is used
var _ = time.Now
