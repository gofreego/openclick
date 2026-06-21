package service

import (
	"encoding/json"
	"net/http"

	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/openclick/internal/models/dao"
	"github.com/gofreego/openclick/internal/models/filter"
)

// ─────────────────────────────────────────────────────────────────────────────
// Dashboards
// ─────────────────────────────────────────────────────────────────────────────

// ListDashboards handles GET /api/v1/projects/:project_id/dashboards
func (s *Service) ListDashboards(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "dashboards:read") {
		writeError(w, http.StatusForbidden, "missing permission: dashboards:read", "FORBIDDEN")
		return
	}
	if !s.assertMembership(ctx, w, projectID, userID) {
		return
	}

	dashboards, counts, err := s.repo.ListDashboards(ctx, &filter.DashboardFilter{ProjectID: projectID})
	if err != nil {
		logger.Error(ctx, "list dashboards: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list dashboards", "INTERNAL_ERROR")
		return
	}

	results := make([]map[string]interface{}, 0, len(dashboards))
	for i, d := range dashboards {
		cnt := 0
		if i < len(counts) {
			cnt = counts[i]
		}
		results = append(results, map[string]interface{}{
			"id":         d.ID,
			"name":       d.Name,
			"item_count": cnt,
			"created_at": d.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"results": results})
}

// CreateDashboard handles POST /api/v1/projects/:project_id/dashboards
func (s *Service) CreateDashboard(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "dashboards:write") {
		writeError(w, http.StatusForbidden, "missing permission: dashboards:write", "FORBIDDEN")
		return
	}
	if !s.assertMembership(ctx, w, projectID, userID) {
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	if body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required", "VALIDATION_ERROR")
		return
	}

	d, err := s.repo.CreateDashboard(ctx, &dao.Dashboard{
		ProjectID: projectID,
		Name:      body.Name,
		Layout:    json.RawMessage("[]"),
	})
	if err != nil {
		logger.Error(ctx, "create dashboard: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create dashboard", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         d.ID,
		"name":       d.Name,
		"created_at": d.CreatedAt,
	})
}

// GetDashboard handles GET /api/v1/projects/:project_id/dashboards/:dashboard_id
func (s *Service) GetDashboard(w http.ResponseWriter, r *http.Request, projectID, dashboardID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "dashboards:read") {
		writeError(w, http.StatusForbidden, "missing permission: dashboards:read", "FORBIDDEN")
		return
	}
	if !s.assertMembership(ctx, w, projectID, userID) {
		return
	}

	d, items, err := s.repo.GetDashboard(ctx, projectID, dashboardID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}

	itemResponses := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		itemResponses = append(itemResponses, dashboardItemToResponse(item))
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         d.ID,
		"name":       d.Name,
		"created_at": d.CreatedAt,
		"items":      itemResponses,
	})
}

// DeleteDashboard handles DELETE /api/v1/projects/:project_id/dashboards/:dashboard_id
func (s *Service) DeleteDashboard(w http.ResponseWriter, r *http.Request, projectID, dashboardID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "dashboards:delete") {
		writeError(w, http.StatusForbidden, "missing permission: dashboards:delete", "FORBIDDEN")
		return
	}
	if !s.assertMembership(ctx, w, projectID, userID) {
		return
	}

	if err := s.repo.DeleteDashboard(ctx, projectID, dashboardID); err != nil {
		writeError(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─────────────────────────────────────────────────────────────────────────────
// Dashboard Items
// ─────────────────────────────────────────────────────────────────────────────

// CreateDashboardItem handles POST /api/v1/projects/:project_id/dashboards/:dashboard_id/items
func (s *Service) CreateDashboardItem(w http.ResponseWriter, r *http.Request, projectID, dashboardID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "dashboards:write") {
		writeError(w, http.StatusForbidden, "missing permission: dashboards:write", "FORBIDDEN")
		return
	}
	if !s.assertMembership(ctx, w, projectID, userID) {
		return
	}

	// Verify dashboard belongs to project
	_, _, err := s.repo.GetDashboard(ctx, projectID, dashboardID)
	if err != nil {
		writeError(w, http.StatusNotFound, "dashboard not found", "NOT_FOUND")
		return
	}

	var body struct {
		Name     string          `json:"name"`
		Type     string          `json:"type"`
		Query    json.RawMessage `json:"query"`
		Position json.RawMessage `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	if body.Name == "" || body.Type == "" {
		writeError(w, http.StatusBadRequest, "name and type are required", "VALIDATION_ERROR")
		return
	}
	if body.Position == nil {
		body.Position = json.RawMessage(`{"x":0,"y":0,"w":6,"h":4}`)
	}

	item, err := s.repo.CreateDashboardItem(ctx, &dao.DashboardItem{
		DashboardID: dashboardID,
		Name:        body.Name,
		Type:        body.Type,
		Query:       body.Query,
		Position:    body.Position,
	})
	if err != nil {
		logger.Error(ctx, "create dashboard item: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create dashboard item", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusCreated, dashboardItemToResponse(item))
}

// UpdateDashboardItem handles PATCH /api/v1/projects/:project_id/dashboards/:dashboard_id/items/:item_id
func (s *Service) UpdateDashboardItem(w http.ResponseWriter, r *http.Request, projectID, dashboardID, itemID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "dashboards:write") {
		writeError(w, http.StatusForbidden, "missing permission: dashboards:write", "FORBIDDEN")
		return
	}
	if !s.assertMembership(ctx, w, projectID, userID) {
		return
	}

	var body struct {
		Name     *string          `json:"name"`
		Type     *string          `json:"type"`
		Query    *json.RawMessage `json:"query"`
		Position *json.RawMessage `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	item := &dao.DashboardItem{ID: itemID, DashboardID: dashboardID}
	if body.Name != nil {
		item.Name = *body.Name
	}
	if body.Type != nil {
		item.Type = *body.Type
	}
	if body.Query != nil {
		item.Query = *body.Query
	}
	if body.Position != nil {
		item.Position = *body.Position
	}

	updated, err := s.repo.UpdateDashboardItem(ctx, item)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}
	writeJSON(w, http.StatusOK, dashboardItemToResponse(updated))
}

// DeleteDashboardItem handles DELETE /api/v1/projects/:project_id/dashboards/:dashboard_id/items/:item_id
func (s *Service) DeleteDashboardItem(w http.ResponseWriter, r *http.Request, projectID, dashboardID, itemID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "dashboards:delete") {
		writeError(w, http.StatusForbidden, "missing permission: dashboards:delete", "FORBIDDEN")
		return
	}
	if !s.assertMembership(ctx, w, projectID, userID) {
		return
	}

	if err := s.repo.DeleteDashboardItem(ctx, dashboardID, itemID); err != nil {
		writeError(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func dashboardItemToResponse(item *dao.DashboardItem) map[string]interface{} {
	return map[string]interface{}{
		"id":           item.ID,
		"dashboard_id": item.DashboardID,
		"name":         item.Name,
		"type":         item.Type,
		"query":        item.Query,
		"position":     item.Position,
	}
}
