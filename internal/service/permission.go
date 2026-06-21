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
		{Key: constants.PermAnalyticsRead, Description: "Read access to analytics"},
		{Key: constants.PermEventsRead, Description: "Read access to events"},
		{Key: constants.PermReplayRead, Description: "Read access to session replays"},
		{Key: constants.PermReplayDelete, Description: "Delete session replays"},
		{Key: constants.PermPersonsRead, Description: "Read access to persons and cohorts"},
		{Key: constants.PermPersonsDelete, Description: "Delete persons and cohorts"},
		{Key: constants.PermFlagsRead, Description: "Read access to feature flags"},
		{Key: constants.PermFlagsWrite, Description: "Create and update feature flags"},
		{Key: constants.PermFlagsDelete, Description: "Delete feature flags"},
	}

	return &openclick_v1.ListPermissionsResponse{
		Permissions: perms,
	}, nil
}
