package service

import (
	"github.com/gofreego/openclick/internal/constants"

	"context"

	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/openclick/api/openclick_v1"
	"github.com/gofreego/openclick/internal/models/dao"
	"github.com/gofreego/openclick/internal/models/filter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ─────────────────────────────────────────────────────────────────────────────
// Persons
// ─────────────────────────────────────────────────────────────────────────────

func (s *Service) ListPersons(ctx context.Context, req *openclick_v1.ListPersonsRequest) (*openclick_v1.ListPersonsResponse, error) {
	if err := s.checkPersonAuth(ctx, req.ProjectId, constants.PermPersonsRead); err != nil {
		return nil, err
	}

	search := ""
	if req.Search != nil {
		search = *req.Search
	}
	limit := 100
	if req.Limit != nil {
		limit = int(*req.Limit)
	}
	offset := 0
	if req.Offset != nil {
		offset = int(*req.Offset)
	}

	f := &filter.PersonFilter{
		ProjectID: req.ProjectId,
		Search:    search,
		Limit:     limit,
		Offset:    offset,
	}

	persons, total, err := s.repo.ListPersons(ctx, f)
	if err != nil {
		logger.Error(ctx, "list persons: %v", err)
		return nil, status.Error(codes.Internal, "failed to list persons")
	}

	var results []*openclick_v1.PersonResponse
	for _, p := range persons {
		results = append(results, personToResponse(p))
	}
	if results == nil {
		results = []*openclick_v1.PersonResponse{}
	}
	return &openclick_v1.ListPersonsResponse{Results: results, Total: int64(total)}, nil
}

func (s *Service) GetPerson(ctx context.Context, req *openclick_v1.GetPersonRequest) (*openclick_v1.GetPersonResponse, error) {
	if err := s.checkPersonAuth(ctx, req.ProjectId, constants.PermPersonsRead); err != nil {
		return nil, err
	}

	person, err := s.repo.GetPerson(ctx, req.ProjectId, req.DistinctId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	resp := &openclick_v1.GetPersonResponse{
		Person: personToResponse(person),
	}

	// Include recent events if ClickHouse is available
	if s.analyticsDB != nil {
		events, err := s.analyticsDB.QueryEvents(ctx, &filter.EventsQuery{
			ProjectID:  req.ProjectId,
			DistinctID: req.DistinctId,
			Limit:      10,
			OrderBy:    "timestamp",
			OrderDir:   "desc",
		})
		if err == nil && events != nil {
			var recentEvents []*openclick_v1.EventResult
			for _, e := range events.Results {
				var props structpb.Struct
				if len(e.Properties) > 0 {
					_ = props.UnmarshalJSON([]byte(e.Properties))
				}
				recentEvents = append(recentEvents, &openclick_v1.EventResult{
					Uuid:       e.UUID,
					Event:      e.Event,
					DistinctId: e.DistinctID,
					Timestamp:  timestamppb.New(e.Timestamp),
					Properties: &props,
				})
			}
			if recentEvents == nil {
				recentEvents = []*openclick_v1.EventResult{}
			}
			resp.RecentEvents = recentEvents
		}
	}

	return resp, nil
}

func (s *Service) DeletePerson(ctx context.Context, req *openclick_v1.DeletePersonRequest) (*openclick_v1.DeletePersonResponse, error) {
	if err := s.checkPersonAuth(ctx, req.ProjectId, constants.PermPersonsDelete); err != nil {
		return nil, err
	}

	if err := s.repo.DeletePerson(ctx, req.ProjectId, req.DistinctId); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &openclick_v1.DeletePersonResponse{}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Cohorts
// ─────────────────────────────────────────────────────────────────────────────

func (s *Service) ListCohorts(ctx context.Context, req *openclick_v1.ListCohortsRequest) (*openclick_v1.ListCohortsResponse, error) {
	if err := s.checkPersonAuth(ctx, req.ProjectId, constants.PermPersonsRead); err != nil {
		return nil, err
	}

	cohorts, err := s.repo.ListCohorts(ctx, &filter.CohortFilter{ProjectID: req.ProjectId})
	if err != nil {
		logger.Error(ctx, "list cohorts: %v", err)
		return nil, status.Error(codes.Internal, "failed to list cohorts")
	}

	var results []*openclick_v1.CohortResponse
	for _, c := range cohorts {
		var filters structpb.Struct
		if len(c.Filters) > 0 {
			_ = filters.UnmarshalJSON(c.Filters)
		}
		results = append(results, &openclick_v1.CohortResponse{
			Id:          c.ID,
			Name:        c.Name,
			Filters:     &filters,
			PersonCount: int64(c.PersonCount),
			CreatedAt:   timestamppb.New(c.CreatedAt),
		})
	}
	if results == nil {
		results = []*openclick_v1.CohortResponse{}
	}
	return &openclick_v1.ListCohortsResponse{Results: results}, nil
}

func (s *Service) CreateCohort(ctx context.Context, req *openclick_v1.CreateCohortRequest) (*openclick_v1.CohortResponse, error) {
	if err := s.checkPersonAuth(ctx, req.ProjectId, constants.PermPersonsRead); err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	filtersBytes, _ := req.Filters.MarshalJSON()
	if len(filtersBytes) == 0 || string(filtersBytes) == "null" {
		filtersBytes = []byte("{}")
	}

	cohort, err := s.repo.CreateCohort(ctx, &dao.Cohort{
		ProjectID: req.ProjectId,
		Name:      req.Name,
		Filters:   filtersBytes,
	})
	if err != nil {
		logger.Error(ctx, "create cohort: %v", err)
		return nil, status.Error(codes.Internal, "failed to create cohort")
	}

	var filters structpb.Struct
	_ = filters.UnmarshalJSON(cohort.Filters)

	return &openclick_v1.CohortResponse{
		Id:          cohort.ID,
		Name:        cohort.Name,
		Filters:     &filters,
		PersonCount: int64(cohort.PersonCount),
		CreatedAt:   timestamppb.New(cohort.CreatedAt),
	}, nil
}

func (s *Service) DeleteCohort(ctx context.Context, req *openclick_v1.DeleteCohortRequest) (*openclick_v1.DeleteCohortResponse, error) {
	if err := s.checkPersonAuth(ctx, req.ProjectId, constants.PermPersonsDelete); err != nil {
		return nil, err
	}

	if err := s.repo.DeleteCohort(ctx, req.ProjectId, req.CohortId); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &openclick_v1.DeleteCohortResponse{}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func (s *Service) checkPersonAuth(ctx context.Context, projectID, perm string) error {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return err
	}
	if !s.hasPermission(ctx, perm) {
		return status.Error(codes.PermissionDenied, "missing permission: "+perm)
	}
	return s.validateMembership(ctx, projectID, userID)
}

func personToResponse(p *dao.Person) *openclick_v1.PersonResponse {
	var props structpb.Struct
	if len(p.Properties) > 0 {
		_ = props.UnmarshalJSON(p.Properties)
	}
	return &openclick_v1.PersonResponse{
		Id:         p.ID,
		DistinctId: p.DistinctID,
		Properties: &props,
		CreatedAt:  timestamppb.New(p.CreatedAt),
	}
}
