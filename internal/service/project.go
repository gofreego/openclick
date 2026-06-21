package service

import (
	"context"
	"strings"

	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/openclick/api/openclick_v1"
	"github.com/gofreego/openclick/internal/models/dao"
	"github.com/gofreego/openclick/internal/models/filter"
	"github.com/gofreego/openclick/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ─────────────────────────────────────────────────────────────────────────────
// ProjectService — all project-related service methods
// ─────────────────────────────────────────────────────────────────────────────

func (s *Service) ListProjects(ctx context.Context, req *openclick_v1.ListProjectsRequest) (*openclick_v1.ListProjectsResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	if !s.hasPermission(ctx, "projects:read") {
		return nil, status.Error(codes.PermissionDenied, "missing permission: projects:read")
	}

	projects, err := s.repo.ListProjectsByUserID(ctx, userID)
	if err != nil {
		logger.Error(ctx, "list projects: %v", err)
		return nil, status.Error(codes.Internal, "failed to list projects")
	}

	var results []*openclick_v1.ProjectResponse
	for _, p := range projects {
		results = append(results, &openclick_v1.ProjectResponse{
			Id:        p.ID,
			Name:      p.Name,
			ApiKey:    p.APIKey,
			SecretKey: p.SecretKey,
			Timezone:  p.Timezone,
			CreatedAt: timestamppb.New(p.CreatedAt),
			UpdatedAt: timestamppb.New(p.UpdatedAt),
		})
	}
	if results == nil {
		results = []*openclick_v1.ProjectResponse{}
	}
	return &openclick_v1.ListProjectsResponse{Results: results}, nil
}

func (s *Service) CreateProject(ctx context.Context, req *openclick_v1.CreateProjectRequest) (*openclick_v1.ProjectResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	if !s.hasPermission(ctx, "projects:write") {
		return nil, status.Error(codes.PermissionDenied, "missing permission: projects:write")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Timezone == "" {
		req.Timezone = "UTC"
	}

	p := &dao.Project{
		Name:      req.Name,
		APIKey:    utils.GeneratePublicKey(),
		SecretKey: utils.GenerateSecretKey(),
		Timezone:  req.Timezone,
	}
	created, err := s.repo.CreateProject(ctx, p)
	if err != nil {
		logger.Error(ctx, "create project: %v", err)
		return nil, status.Error(codes.Internal, "failed to create project")
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

	return &openclick_v1.ProjectResponse{
		Id:        created.ID,
		Name:      created.Name,
		ApiKey:    created.APIKey,
		SecretKey: created.SecretKey,
		Timezone:  created.Timezone,
		CreatedAt: timestamppb.New(created.CreatedAt),
		UpdatedAt: timestamppb.New(created.UpdatedAt),
	}, nil
}

func (s *Service) GetProject(ctx context.Context, req *openclick_v1.GetProjectRequest) (*openclick_v1.GetProjectResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	if !s.hasPermission(ctx, "projects:read") {
		return nil, status.Error(codes.PermissionDenied, "missing permission: projects:read")
	}

	ok, _ := s.repo.IsProjectMember(ctx, req.ProjectId, userID)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "not a member of this project")
	}

	project, err := s.repo.GetProjectByID(ctx, req.ProjectId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	members, err := s.repo.GetProjectMembers(ctx, req.ProjectId)
	if err != nil {
		logger.Error(ctx, "get project members: %v", err)
	}

	var memberList []*openclick_v1.ProjectMemberResponse
	for _, m := range members {
		memberList = append(memberList, &openclick_v1.ProjectMemberResponse{
			UserId:    m.UserID,
			Role:      m.Role,
			CreatedAt: timestamppb.New(m.CreatedAt),
		})
	}
	if memberList == nil {
		memberList = []*openclick_v1.ProjectMemberResponse{}
	}

	return &openclick_v1.GetProjectResponse{
		Id:        project.ID,
		Name:      project.Name,
		ApiKey:    project.APIKey,
		SecretKey: project.SecretKey,
		Timezone:  project.Timezone,
		CreatedAt: timestamppb.New(project.CreatedAt),
		Members:   memberList,
	}, nil
}

func (s *Service) UpdateProject(ctx context.Context, req *openclick_v1.UpdateProjectRequest) (*openclick_v1.ProjectResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	if !s.hasPermission(ctx, "projects:write") {
		return nil, status.Error(codes.PermissionDenied, "missing permission: projects:write")
	}
	ok, _ := s.repo.IsProjectMember(ctx, req.ProjectId, userID)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "not a member of this project")
	}

	project, err := s.repo.GetProjectByID(ctx, req.ProjectId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Timezone != nil {
		project.Timezone = *req.Timezone
	}

	updated, err := s.repo.UpdateProject(ctx, project)
	if err != nil {
		logger.Error(ctx, "update project: %v", err)
		return nil, status.Error(codes.Internal, "failed to update project")
	}
	return &openclick_v1.ProjectResponse{
		Id:        updated.ID,
		Name:      updated.Name,
		ApiKey:    updated.APIKey,
		SecretKey: updated.SecretKey,
		Timezone:  updated.Timezone,
		CreatedAt: timestamppb.New(updated.CreatedAt),
		UpdatedAt: timestamppb.New(updated.UpdatedAt),
	}, nil
}

func (s *Service) DeleteProject(ctx context.Context, req *openclick_v1.DeleteProjectRequest) (*openclick_v1.DeleteProjectResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	if !s.hasPermission(ctx, "projects:delete") {
		return nil, status.Error(codes.PermissionDenied, "missing permission: projects:delete")
	}
	if err := s.validateMembership(ctx, req.ProjectId, userID); err != nil {
		return nil, err
	}
	if err := s.repo.DeleteProject(ctx, req.ProjectId); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &openclick_v1.DeleteProjectResponse{}, nil
}

func (s *Service) AddMember(ctx context.Context, req *openclick_v1.AddMemberRequest) (*openclick_v1.AddMemberResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	if !s.hasPermission(ctx, "members:write") {
		return nil, status.Error(codes.PermissionDenied, "missing permission: members:write")
	}
	ok, _ := s.repo.IsProjectMember(ctx, req.ProjectId, userID)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "not a member of this project")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.Role == "" {
		req.Role = "member"
	}

	m, err := s.repo.AddProjectMember(ctx, &dao.ProjectMember{
		ProjectID: req.ProjectId,
		UserID:    req.UserId,
		Role:      req.Role,
	})
	if err != nil {
		logger.Error(ctx, "add project member: %v", err)
		return nil, status.Error(codes.Internal, "failed to add member")
	}
	return &openclick_v1.AddMemberResponse{
		UserId:    m.UserID,
		Role:      m.Role,
		CreatedAt: timestamppb.New(m.CreatedAt),
	}, nil
}

func (s *Service) RemoveMember(ctx context.Context, req *openclick_v1.RemoveMemberRequest) (*openclick_v1.RemoveMemberResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	if !s.hasPermission(ctx, "members:write") {
		return nil, status.Error(codes.PermissionDenied, "missing permission: members:write")
	}
	ok, _ := s.repo.IsProjectMember(ctx, req.ProjectId, userID)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "not a member of this project")
	}
	if err := s.repo.RemoveProjectMember(ctx, req.ProjectId, req.UserId); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &openclick_v1.RemoveMemberResponse{}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func (s *Service) getUserID(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}
	ids := md.Get("x-user-id")
	if len(ids) == 0 || ids[0] == "" {
		return "", status.Error(codes.Unauthenticated, "x-user-id header is required")
	}
	return ids[0], nil
}

func (s *Service) hasPermission(ctx context.Context, scope string) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	perms := md.Get("x-user-perms")
	if len(perms) == 0 {
		return false
	}
	for _, p := range strings.Split(perms[0], ",") {
		if strings.TrimSpace(p) == scope {
			return true
		}
	}
	return false
}

// validateMembership checks that projectID is valid and userID is a member
func (s *Service) validateMembership(ctx context.Context, projectID, userID string) error {
	ok, err := s.repo.IsProjectMember(ctx, projectID, userID)
	if err != nil || !ok {
		return status.Error(codes.PermissionDenied, "not a member of this project")
	}
	return nil
}

// Ensure unused imports are kept
var _ = filter.ProjectFilter{}
