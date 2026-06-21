package service

import (
	"context"
	"encoding/json"

	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/openclick/api/openclick_v1"
	"github.com/gofreego/openclick/internal/constants"
	"github.com/gofreego/openclick/internal/models/dao"
	"github.com/gofreego/openclick/internal/models/filter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ─────────────────────────────────────────────────────────────────────────────
// Dashboards
// ─────────────────────────────────────────────────────────────────────────────

func (s *Service) ListDashboards(ctx context.Context, req *openclick_v1.ListDashboardsRequest) (*openclick_v1.ListDashboardsResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	if !s.hasPermission(ctx, constants.PermDashboardsRead) {
		return nil, status.Error(codes.PermissionDenied, "missing permission: dashboards:read")
	}
	if err := s.validateMembership(ctx, req.ProjectId, userID); err != nil {
		return nil, err
	}

	dashboards, counts, err := s.repo.ListDashboards(ctx, &filter.DashboardFilter{ProjectID: req.ProjectId})
	if err != nil {
		logger.Error(ctx, "list dashboards: %v", err)
		return nil, status.Error(codes.Internal, "failed to list dashboards")
	}

	var results []*openclick_v1.DashboardResponse
	for i, d := range dashboards {
		cnt := int32(0)
		if i < len(counts) {
			cnt = int32(counts[i])
		}
		results = append(results, &openclick_v1.DashboardResponse{
			Id:        d.ID,
			Name:      d.Name,
			ItemCount: cnt,
			CreatedAt: timestamppb.New(d.CreatedAt),
		})
	}
	if results == nil {
		results = []*openclick_v1.DashboardResponse{}
	}
	return &openclick_v1.ListDashboardsResponse{Results: results}, nil
}

func (s *Service) CreateDashboard(ctx context.Context, req *openclick_v1.CreateDashboardRequest) (*openclick_v1.DashboardResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	if !s.hasPermission(ctx, constants.PermDashboardsWrite) {
		return nil, status.Error(codes.PermissionDenied, "missing permission: dashboards:write")
	}
	if err := s.validateMembership(ctx, req.ProjectId, userID); err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	d, err := s.repo.CreateDashboard(ctx, &dao.Dashboard{
		ProjectID: req.ProjectId,
		Name:      req.Name,
		Layout:    json.RawMessage("[]"),
	})
	if err != nil {
		logger.Error(ctx, "create dashboard: %v", err)
		return nil, status.Error(codes.Internal, "failed to create dashboard")
	}

	return &openclick_v1.DashboardResponse{
		Id:        d.ID,
		Name:      d.Name,
		ItemCount: 0,
		CreatedAt: timestamppb.New(d.CreatedAt),
	}, nil
}

func (s *Service) GetDashboard(ctx context.Context, req *openclick_v1.GetDashboardRequest) (*openclick_v1.GetDashboardResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	if !s.hasPermission(ctx, constants.PermDashboardsRead) {
		return nil, status.Error(codes.PermissionDenied, "missing permission: dashboards:read")
	}
	if err := s.validateMembership(ctx, req.ProjectId, userID); err != nil {
		return nil, err
	}

	d, items, err := s.repo.GetDashboard(ctx, req.ProjectId, req.DashboardId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	var itemResponses []*openclick_v1.DashboardItemResponse
	for _, item := range items {
		itemResponses = append(itemResponses, dashboardItemToResponse(item))
	}
	if itemResponses == nil {
		itemResponses = []*openclick_v1.DashboardItemResponse{}
	}

	return &openclick_v1.GetDashboardResponse{
		Id:        d.ID,
		Name:      d.Name,
		CreatedAt: timestamppb.New(d.CreatedAt),
		Items:     itemResponses,
	}, nil
}

func (s *Service) DeleteDashboard(ctx context.Context, req *openclick_v1.DeleteDashboardRequest) (*openclick_v1.DeleteDashboardResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	if !s.hasPermission(ctx, constants.PermDashboardsDelete) {
		return nil, status.Error(codes.PermissionDenied, "missing permission: dashboards:delete")
	}
	if err := s.validateMembership(ctx, req.ProjectId, userID); err != nil {
		return nil, err
	}

	if err := s.repo.DeleteDashboard(ctx, req.ProjectId, req.DashboardId); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &openclick_v1.DeleteDashboardResponse{}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Dashboard Items
// ─────────────────────────────────────────────────────────────────────────────

func (s *Service) CreateDashboardItem(ctx context.Context, req *openclick_v1.CreateDashboardItemRequest) (*openclick_v1.DashboardItemResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	if !s.hasPermission(ctx, constants.PermDashboardsWrite) {
		return nil, status.Error(codes.PermissionDenied, "missing permission: dashboards:write")
	}
	if err := s.validateMembership(ctx, req.ProjectId, userID); err != nil {
		return nil, err
	}

	// Verify dashboard belongs to project
	_, _, err = s.repo.GetDashboard(ctx, req.ProjectId, req.DashboardId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "dashboard not found")
	}

	if req.Name == "" || req.Type == "" {
		return nil, status.Error(codes.InvalidArgument, "name and type are required")
	}

	queryBytes, _ := req.Query.MarshalJSON()
	var positionBytes []byte
	if req.Position != nil {
		positionBytes, _ = req.Position.MarshalJSON()
	} else {
		positionBytes = []byte(`{"x":0,"y":0,"w":6,"h":4}`)
	}

	item, err := s.repo.CreateDashboardItem(ctx, &dao.DashboardItem{
		DashboardID: req.DashboardId,
		Name:        req.Name,
		Type:        req.Type,
		Query:       json.RawMessage(queryBytes),
		Position:    json.RawMessage(positionBytes),
	})
	if err != nil {
		logger.Error(ctx, "create dashboard item: %v", err)
		return nil, status.Error(codes.Internal, "failed to create dashboard item")
	}

	return dashboardItemToResponse(item), nil
}

func (s *Service) UpdateDashboardItem(ctx context.Context, req *openclick_v1.UpdateDashboardItemRequest) (*openclick_v1.DashboardItemResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	if !s.hasPermission(ctx, constants.PermDashboardsWrite) {
		return nil, status.Error(codes.PermissionDenied, "missing permission: dashboards:write")
	}
	if err := s.validateMembership(ctx, req.ProjectId, userID); err != nil {
		return nil, err
	}

	item := &dao.DashboardItem{ID: req.ItemId, DashboardID: req.DashboardId}
	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Type != nil {
		item.Type = *req.Type
	}
	if req.Query != nil {
		b, _ := req.Query.MarshalJSON()
		item.Query = json.RawMessage(b)
	}
	if req.Position != nil {
		b, _ := req.Position.MarshalJSON()
		item.Position = json.RawMessage(b)
	}

	updated, err := s.repo.UpdateDashboardItem(ctx, item)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return dashboardItemToResponse(updated), nil
}

func (s *Service) DeleteDashboardItem(ctx context.Context, req *openclick_v1.DeleteDashboardItemRequest) (*openclick_v1.DeleteDashboardItemResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	if !s.hasPermission(ctx, constants.PermDashboardsDelete) {
		return nil, status.Error(codes.PermissionDenied, "missing permission: dashboards:delete")
	}
	if err := s.validateMembership(ctx, req.ProjectId, userID); err != nil {
		return nil, err
	}

	if err := s.repo.DeleteDashboardItem(ctx, req.DashboardId, req.ItemId); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &openclick_v1.DeleteDashboardItemResponse{}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func dashboardItemToResponse(item *dao.DashboardItem) *openclick_v1.DashboardItemResponse {
	var queryStruct structpb.Struct
	if len(item.Query) > 0 {
		_ = queryStruct.UnmarshalJSON(item.Query)
	}
	var posStruct structpb.Struct
	if len(item.Position) > 0 {
		_ = posStruct.UnmarshalJSON(item.Position)
	}
	return &openclick_v1.DashboardItemResponse{
		Id:          item.ID,
		DashboardId: item.DashboardID,
		Name:        item.Name,
		Type:        item.Type,
		Query:       &queryStruct,
		Position:    &posStruct,
	}
}
