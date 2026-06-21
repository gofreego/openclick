package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/openclick/api/openclick_v1"
	"github.com/gofreego/openclick/internal/models/dao"
	"github.com/gofreego/openclick/internal/models/filter"
	"github.com/gofreego/openclick/pkg/utils"
)

// ─────────────────────────────────────────────────────────────────────────────
// ProjectService — all project-related service methods
// ─────────────────────────────────────────────────────────────────────────────

// ListProjects handles GET /api/v1/projects
func (s *Service) ListProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "projects:read") {
		writeError(w, http.StatusForbidden, "missing permission: projects:read", "FORBIDDEN")
		return
	}

	projects, err := s.repo.ListProjectsByUserID(ctx, userID)
	if err != nil {
		logger.Error(ctx, "list projects: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list projects", "INTERNAL_ERROR")
		return
	}

	type projectResponse struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		APIKey    string    `json:"api_key"`
		Timezone  string    `json:"timezone"`
		CreatedAt time.Time `json:"created_at"`
	}
	var results []projectResponse
	for _, p := range projects {
		results = append(results, projectResponse{
			ID:        p.ID,
			Name:      p.Name,
			APIKey:    p.APIKey,
			Timezone:  p.Timezone,
			CreatedAt: p.CreatedAt,
		})
	}
	if results == nil {
		results = []projectResponse{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"results": results})
}

// CreateProject handles POST /api/v1/projects
func (s *Service) CreateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "projects:write") {
		writeError(w, http.StatusForbidden, "missing permission: projects:write", "FORBIDDEN")
		return
	}

	var body struct {
		Name     string `json:"name"`
		Timezone string `json:"timezone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	if body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required", "VALIDATION_ERROR")
		return
	}
	if body.Timezone == "" {
		body.Timezone = "UTC"
	}

	p := &dao.Project{
		Name:      body.Name,
		APIKey:    utils.GeneratePublicKey(),
		SecretKey: utils.GenerateSecretKey(),
		Timezone:  body.Timezone,
	}
	created, err := s.repo.CreateProject(ctx, p)
	if err != nil {
		logger.Error(ctx, "create project: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create project", "INTERNAL_ERROR")
		return
	}

	// Auto-add creator as owner
	_, err = s.repo.AddProjectMember(ctx, &dao.ProjectMember{
		ProjectID: created.ID,
		UserID:    userID,
		Role:      "owner",
	})
	if err != nil {
		logger.Error(ctx, "add project owner: %v", err)
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         created.ID,
		"name":       created.Name,
		"api_key":    created.APIKey,
		"secret_key": created.SecretKey,
		"timezone":   created.Timezone,
		"created_at": created.CreatedAt,
	})
}

// GetProject handles GET /api/v1/projects/:project_id
func (s *Service) GetProject(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "projects:read") {
		writeError(w, http.StatusForbidden, "missing permission: projects:read", "FORBIDDEN")
		return
	}

	ok, _ := s.repo.IsProjectMember(ctx, projectID, userID)
	if !ok {
		writeError(w, http.StatusForbidden, "not a member of this project", "FORBIDDEN")
		return
	}

	project, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}

	members, err := s.repo.GetProjectMembers(ctx, projectID)
	if err != nil {
		logger.Error(ctx, "get project members: %v", err)
	}

	type memberResponse struct {
		UserID    string    `json:"user_id"`
		Role      string    `json:"role"`
		CreatedAt time.Time `json:"created_at"`
	}
	var memberList []memberResponse
	for _, m := range members {
		memberList = append(memberList, memberResponse{UserID: m.UserID, Role: m.Role, CreatedAt: m.CreatedAt})
	}
	if memberList == nil {
		memberList = []memberResponse{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         project.ID,
		"name":       project.Name,
		"api_key":    project.APIKey,
		"secret_key": project.SecretKey,
		"timezone":   project.Timezone,
		"created_at": project.CreatedAt,
		"members":    memberList,
	})
}

// UpdateProject handles PATCH /api/v1/projects/:project_id
func (s *Service) UpdateProject(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "projects:write") {
		writeError(w, http.StatusForbidden, "missing permission: projects:write", "FORBIDDEN")
		return
	}
	ok, _ := s.repo.IsProjectMember(ctx, projectID, userID)
	if !ok {
		writeError(w, http.StatusForbidden, "not a member of this project", "FORBIDDEN")
		return
	}

	project, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}

	var body struct {
		Name     *string `json:"name"`
		Timezone *string `json:"timezone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	if body.Name != nil {
		project.Name = *body.Name
	}
	if body.Timezone != nil {
		project.Timezone = *body.Timezone
	}

	updated, err := s.repo.UpdateProject(ctx, project)
	if err != nil {
		logger.Error(ctx, "update project: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to update project", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         updated.ID,
		"name":       updated.Name,
		"api_key":    updated.APIKey,
		"secret_key": updated.SecretKey,
		"timezone":   updated.Timezone,
		"created_at": updated.CreatedAt,
		"updated_at": updated.UpdatedAt,
	})
}

// DeleteProject handles DELETE /api/v1/projects/:project_id
func (s *Service) DeleteProject(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "projects:delete") {
		writeError(w, http.StatusForbidden, "missing permission: projects:delete", "FORBIDDEN")
		return
	}
	ok, _ := s.repo.IsProjectMember(ctx, projectID, userID)
	if !ok {
		writeError(w, http.StatusForbidden, "not a member of this project", "FORBIDDEN")
		return
	}
	if err := s.repo.DeleteProject(ctx, projectID); err != nil {
		writeError(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AddMember handles POST /api/v1/projects/:project_id/members
func (s *Service) AddMember(w http.ResponseWriter, r *http.Request, projectID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "members:write") {
		writeError(w, http.StatusForbidden, "missing permission: members:write", "FORBIDDEN")
		return
	}
	ok, _ := s.repo.IsProjectMember(ctx, projectID, userID)
	if !ok {
		writeError(w, http.StatusForbidden, "not a member of this project", "FORBIDDEN")
		return
	}

	var body struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	if body.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required", "VALIDATION_ERROR")
		return
	}
	if body.Role == "" {
		body.Role = "member"
	}

	m, err := s.repo.AddProjectMember(ctx, &dao.ProjectMember{
		ProjectID: projectID,
		UserID:    body.UserID,
		Role:      body.Role,
	})
	if err != nil {
		logger.Error(ctx, "add project member: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to add member", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":    m.UserID,
		"role":       m.Role,
		"created_at": m.CreatedAt,
	})
}

// RemoveMember handles DELETE /api/v1/projects/:project_id/members/:user_id
func (s *Service) RemoveMember(w http.ResponseWriter, r *http.Request, projectID, targetUserID string) {
	ctx := r.Context()
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "x-user-id header is required", "UNAUTHORIZED")
		return
	}
	if !hasPermission(r, "members:write") {
		writeError(w, http.StatusForbidden, "missing permission: members:write", "FORBIDDEN")
		return
	}
	ok, _ := s.repo.IsProjectMember(ctx, projectID, userID)
	if !ok {
		writeError(w, http.StatusForbidden, "not a member of this project", "FORBIDDEN")
		return
	}
	if err := s.repo.RemoveProjectMember(ctx, projectID, targetUserID); err != nil {
		writeError(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// hasPermission checks if x-user-perms header contains the given scope
func hasPermission(r *http.Request, scope string) bool {
	perms := r.Header.Get("x-user-perms")
	for _, p := range strings.Split(perms, ",") {
		if strings.TrimSpace(p) == scope {
			return true
		}
	}
	return false
}

// writeJSON encodes val as JSON and writes it to the ResponseWriter
func writeJSON(w http.ResponseWriter, status int, val interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(val)
}

// writeError writes a standard error JSON response
func writeError(w http.ResponseWriter, status int, message, code string) {
	writeJSON(w, status, map[string]string{"error": message, "code": code})
}

// assertMembership checks that projectID is valid and userID is a member — returns false on failure
func (s *Service) assertMembership(ctx context.Context, w http.ResponseWriter, projectID, userID string) bool {
	ok, err := s.repo.IsProjectMember(ctx, projectID, userID)
	if err != nil || !ok {
		writeError(w, http.StatusForbidden, "not a member of this project", "FORBIDDEN")
		return false
	}
	return true
}

// Ensure unused imports are kept (filter is used indirectly via type reference in interface)
var _ = filter.ProjectFilter{}
var _ = fmt.Sprintf
var _ = openclick_v1.UnimplementedBaseServiceServer{}
