package service

import (
	"context"

	"github.com/gofreego/openclick/api/openclick_v1"
	"github.com/gofreego/openclick/internal/constants"
)

func (s *Service) ListPermissions(ctx context.Context, req *openclick_v1.ListPermissionsRequest) (*openclick_v1.ListPermissionsResponse, error) {
	// Static list of all available permissions in the system
	perms := []*openclick_v1.Permission{
		{Key: constants.PermProjectsRead, Description: "Read access to projects"},
		{Key: constants.PermProjectsWrite, Description: "Create and update projects"},
		{Key: constants.PermProjectsDelete, Description: "Delete projects"},
		{Key: constants.PermDashboardsRead, Description: "Read access to dashboards"},
		{Key: constants.PermDashboardsWrite, Description: "Create and update dashboards"},
		{Key: constants.PermDashboardsDelete, Description: "Delete dashboards"},
		{Key: constants.PermMembersWrite, Description: "Manage project members"},
	}

	return &openclick_v1.ListPermissionsResponse{
		Permissions: perms,
	}, nil
}
