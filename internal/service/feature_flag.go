package service

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/openclick/internal/models/dao"
	"github.com/gofreego/openclick/internal/models/filter"
)

// ─────────────────────────────────────────────────────────────────────────────
// Feature Flags
// ─────────────────────────────────────────────────────────────────────────────

// ListFeatureFlags handles GET /api/v1/projects/:project_id/feature-flags
func (s *Service) ListFeatureFlags(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "flags:read") {
		writeError(w, http.StatusForbidden, "missing permission: flags:read", "FORBIDDEN")
		return
	}
	if !s.assertMembership(ctx, w, projectID, userID) {
		return
	}

	flags, err := s.repo.ListFeatureFlags(ctx, &filter.FeatureFlagFilter{ProjectID: projectID})
	if err != nil {
		logger.Error(ctx, "list feature flags: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list feature flags", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"results": flagsToResponse(flags)})
}

// CreateFeatureFlag handles POST /api/v1/projects/:project_id/feature-flags
func (s *Service) CreateFeatureFlag(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "flags:write") {
		writeError(w, http.StatusForbidden, "missing permission: flags:write", "FORBIDDEN")
		return
	}
	if !s.assertMembership(ctx, w, projectID, userID) {
		return
	}

	var body struct {
		Key        string          `json:"key"`
		Name       string          `json:"name"`
		Active     bool            `json:"active"`
		RolloutPct int16           `json:"rollout_pct"`
		Filters    json.RawMessage `json:"filters"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	if body.Key == "" || body.Name == "" {
		writeError(w, http.StatusBadRequest, "key and name are required", "VALIDATION_ERROR")
		return
	}
	if body.Filters == nil {
		body.Filters = json.RawMessage("{}")
	}

	f := &dao.FeatureFlag{
		ProjectID:  projectID,
		Key:        body.Key,
		Name:       body.Name,
		Active:     body.Active,
		RolloutPct: body.RolloutPct,
		Filters:    body.Filters,
	}
	created, err := s.repo.CreateFeatureFlag(ctx, f)
	if err != nil {
		logger.Error(ctx, "create feature flag: %v", err)
		writeError(w, http.StatusConflict, err.Error(), "CONFLICT")
		return
	}
	writeJSON(w, http.StatusCreated, flagToResponse(created))
}

// UpdateFeatureFlag handles PATCH /api/v1/projects/:project_id/feature-flags/:flag_id
func (s *Service) UpdateFeatureFlag(w http.ResponseWriter, r *http.Request, projectID, flagID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "flags:write") {
		writeError(w, http.StatusForbidden, "missing permission: flags:write", "FORBIDDEN")
		return
	}
	if !s.assertMembership(ctx, w, projectID, userID) {
		return
	}

	existing, err := s.repo.GetFeatureFlagByID(ctx, projectID, flagID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}

	var body struct {
		Name       *string          `json:"name"`
		Active     *bool            `json:"active"`
		RolloutPct *int16           `json:"rollout_pct"`
		Filters    *json.RawMessage `json:"filters"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	if body.Name != nil {
		existing.Name = *body.Name
	}
	if body.Active != nil {
		existing.Active = *body.Active
	}
	if body.RolloutPct != nil {
		existing.RolloutPct = *body.RolloutPct
	}
	if body.Filters != nil {
		existing.Filters = *body.Filters
	}

	updated, err := s.repo.UpdateFeatureFlag(ctx, existing)
	if err != nil {
		logger.Error(ctx, "update feature flag: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to update feature flag", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusOK, flagToResponse(updated))
}

// DeleteFeatureFlag handles DELETE /api/v1/projects/:project_id/feature-flags/:flag_id
func (s *Service) DeleteFeatureFlag(w http.ResponseWriter, r *http.Request, projectID, flagID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "flags:delete") {
		writeError(w, http.StatusForbidden, "missing permission: flags:delete", "FORBIDDEN")
		return
	}
	if !s.assertMembership(ctx, w, projectID, userID) {
		return
	}

	if err := s.repo.DeleteFeatureFlag(ctx, projectID, flagID); err != nil {
		writeError(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// EvaluateFlags handles POST /api/v1/projects/:project_id/feature-flags/evaluate
// Authenticated via project secret_key Bearer token
func (s *Service) EvaluateFlags(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	secretKey := extractBearerToken(r)
	if secretKey == "" {
		writeError(w, http.StatusUnauthorized, "Bearer secret_key is required", "UNAUTHORIZED")
		return
	}

	project, err := s.repo.GetProjectBySecretKey(ctx, secretKey)
	if err != nil || project.ID != projectID {
		writeError(w, http.StatusUnauthorized, "invalid secret_key", "UNAUTHORIZED")
		return
	}

	var body struct {
		DistinctID       string                 `json:"distinct_id"`
		PersonProperties map[string]interface{} `json:"person_properties"`
		Groups           map[string]interface{} `json:"groups"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	if body.DistinctID == "" {
		writeError(w, http.StatusBadRequest, "distinct_id is required", "VALIDATION_ERROR")
		return
	}

	flags, err := s.repo.ListActiveFeatureFlags(ctx, projectID)
	if err != nil {
		logger.Error(ctx, "evaluate flags: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to evaluate flags", "INTERNAL_ERROR")
		return
	}

	results := evaluateFlags(flags, body.DistinctID, body.PersonProperties)
	writeJSON(w, http.StatusOK, map[string]interface{}{"feature_flags": results})
}

// Decide handles GET /decide/ — client-side evaluation via api_key
func (s *Service) Decide(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	apiKey := r.URL.Query().Get("api_key")
	if apiKey == "" {
		apiKey = extractBearerToken(r)
	}
	distinctID := r.URL.Query().Get("distinct_id")

	if apiKey == "" || distinctID == "" {
		writeError(w, http.StatusBadRequest, "api_key and distinct_id are required", "BAD_REQUEST")
		return
	}

	project, err := s.repo.GetProjectByAPIKey(ctx, apiKey)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid api_key", "UNAUTHORIZED")
		return
	}

	flags, err := s.repo.ListActiveFeatureFlags(ctx, project.ID)
	if err != nil {
		logger.Error(ctx, "decide: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to evaluate flags", "INTERNAL_ERROR")
		return
	}

	results := evaluateFlags(flags, distinctID, nil)
	writeJSON(w, http.StatusOK, map[string]interface{}{"feature_flags": results})
}

// ─────────────────────────────────────────────────────────────────────────────
// Flag evaluation helpers
// ─────────────────────────────────────────────────────────────────────────────

// evaluateFlags determines which flags are enabled for a given user
func evaluateFlags(flags []*dao.FeatureFlag, distinctID string, personProps map[string]interface{}) map[string]bool {
	results := make(map[string]bool, len(flags))
	for _, f := range flags {
		results[f.Key] = isFlagEnabled(f, distinctID, personProps)
	}
	return results
}

// isFlagEnabled determines if a specific flag is enabled for a user
// Uses deterministic hashing on distinctID to split traffic consistently
func isFlagEnabled(f *dao.FeatureFlag, distinctID string, personProps map[string]interface{}) bool {
	if !f.Active {
		return false
	}

	// Check property filter conditions
	var filterCfg struct {
		Groups []struct {
			Properties []struct {
				Key      string      `json:"key"`
				Value    interface{} `json:"value"`
				Operator string      `json:"operator"`
				Type     string      `json:"type"`
			} `json:"properties"`
			RolloutPercentage int `json:"rollout_percentage"`
		} `json:"groups"`
	}
	if len(f.Filters) > 0 && string(f.Filters) != "{}" {
		json.Unmarshal(f.Filters, &filterCfg)
		if len(filterCfg.Groups) > 0 && personProps != nil {
			matchedAny := false
			for _, group := range filterCfg.Groups {
				if matchesGroup(group.Properties, personProps) {
					matchedAny = true
					break
				}
			}
			if !matchedAny {
				return false
			}
		}
	}

	// Rollout percentage check using deterministic hash
	if f.RolloutPct >= 100 {
		return true
	}
	if f.RolloutPct <= 0 {
		return false
	}
	hash := sha256.Sum256([]byte(f.Key + "." + distinctID))
	bucket := int(hash[0]) % 100
	return bucket < int(f.RolloutPct)
}

// matchesGroup checks if personProps satisfies all property conditions in a group
func matchesGroup(conditions []struct {
	Key      string      `json:"key"`
	Value    interface{} `json:"value"`
	Operator string      `json:"operator"`
	Type     string      `json:"type"`
}, personProps map[string]interface{}) bool {
	for _, cond := range conditions {
		val, exists := personProps[cond.Key]
		if !exists {
			return false
		}
		switch cond.Operator {
		case "exact":
			if fmt.Sprintf("%v", val) != fmt.Sprintf("%v", cond.Value) {
				return false
			}
		case "contains":
			if !strings.Contains(strings.ToLower(fmt.Sprintf("%v", val)), strings.ToLower(fmt.Sprintf("%v", cond.Value))) {
				return false
			}
		}
	}
	return true
}

// flagToResponse converts a FeatureFlag DAO to an API response map
func flagToResponse(f *dao.FeatureFlag) map[string]interface{} {
	return map[string]interface{}{
		"id":          f.ID,
		"key":         f.Key,
		"name":        f.Name,
		"active":      f.Active,
		"rollout_pct": f.RolloutPct,
		"filters":     f.Filters,
		"created_at":  f.CreatedAt,
	}
}

func flagsToResponse(flags []*dao.FeatureFlag) []map[string]interface{} {
	if flags == nil {
		return []map[string]interface{}{}
	}
	result := make([]map[string]interface{}, len(flags))
	for i, f := range flags {
		result[i] = flagToResponse(f)
	}
	return result
}
