package service

import (
	"github.com/gofreego/openclick/internal/constants"

	"context"

	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/openclick/api/openclick_v1"
	"github.com/gofreego/openclick/internal/models/filter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ─────────────────────────────────────────────────────────────────────────────
// Analytics Query Handlers
// ─────────────────────────────────────────────────────────────────────────────

func (s *Service) QueryTrends(ctx context.Context, req *openclick_v1.QueryTrendsRequest) (*openclick_v1.QueryTrendsResponse, error) {
	if err := s.checkAnalyticsAuth(ctx, req.ProjectId, constants.PermAnalyticsRead); err != nil {
		return nil, err
	}

	if s.analyticsDB == nil {
		return &openclick_v1.QueryTrendsResponse{Results: []*openclick_v1.TrendsSeries{}}, nil
	}

	var events []filter.TrendsEvent
	for _, e := range req.Events {
		events = append(events, filter.TrendsEvent{ID: e.Id, Name: e.Name, Math: e.Math})
	}

	result, err := s.analyticsDB.QueryTrends(ctx, &filter.TrendsQuery{
		ProjectID: req.ProjectId,
		Events:    events,
		DateFrom:  req.DateFrom,
		DateTo:    req.DateTo,
		Interval:  req.Interval,
		Filters:   protoToPropertyFilters(req.Filters),
		Breakdown: req.Breakdown,
	})
	if err != nil {
		logger.Error(ctx, "query trends: %v", err)
		return nil, status.Error(codes.Internal, "failed to query trends")
	}

	var results []*openclick_v1.TrendsSeries
	for _, series := range result.Results {
		results = append(results, &openclick_v1.TrendsSeries{
			Label:          series.Label,
			BreakdownValue: series.BreakdownValue,
			Data:           series.Data,
			Labels:         series.Labels,
			Days:           series.Days,
		})
	}
	if results == nil {
		results = []*openclick_v1.TrendsSeries{}
	}
	return &openclick_v1.QueryTrendsResponse{Results: results}, nil
}

func (s *Service) QueryFunnel(ctx context.Context, req *openclick_v1.QueryFunnelRequest) (*openclick_v1.QueryFunnelResponse, error) {
	if err := s.checkAnalyticsAuth(ctx, req.ProjectId, constants.PermAnalyticsRead); err != nil {
		return nil, err
	}

	if s.analyticsDB == nil {
		return &openclick_v1.QueryFunnelResponse{Result: []*openclick_v1.FunnelStepResult{}}, nil
	}

	var steps []filter.FunnelStep
	for _, st := range req.Steps {
		steps = append(steps, filter.FunnelStep{Event: st.Event, Name: st.Name})
	}

	result, err := s.analyticsDB.QueryFunnel(ctx, &filter.FunnelQuery{
		ProjectID:            req.ProjectId,
		Steps:                steps,
		DateFrom:             req.DateFrom,
		DateTo:               req.DateTo,
		ConversionWindowDays: int(req.ConversionWindowDays),
		FunnelOrder:          req.FunnelOrder,
		Filters:              protoToPropertyFilters(req.Filters),
	})
	if err != nil {
		logger.Error(ctx, "query funnel: %v", err)
		return nil, status.Error(codes.Internal, "failed to query funnel")
	}

	var resSteps []*openclick_v1.FunnelStepResult
	for _, step := range result.Result {
		avgTime := float64(0)
		if step.AverageConversionTime != nil {
			avgTime = *step.AverageConversionTime
		}
		resSteps = append(resSteps, &openclick_v1.FunnelStepResult{
			ActionId:              step.ActionID,
			Name:                  step.Name,
			Count:                 step.Count,
			ConversionRate:        step.ConversionRate,
			AverageConversionTime: avgTime,
		})
	}
	if resSteps == nil {
		resSteps = []*openclick_v1.FunnelStepResult{}
	}
	return &openclick_v1.QueryFunnelResponse{Result: resSteps}, nil
}

func (s *Service) QueryRetention(ctx context.Context, req *openclick_v1.QueryRetentionRequest) (*openclick_v1.QueryRetentionResponse, error) {
	if err := s.checkAnalyticsAuth(ctx, req.ProjectId, constants.PermAnalyticsRead); err != nil {
		return nil, err
	}

	if s.analyticsDB == nil {
		return &openclick_v1.QueryRetentionResponse{Result: []*openclick_v1.RetentionCohort{}}, nil
	}

	var targetEvent, returnEvent filter.RetentionEvent
	if req.TargetEvent != nil {
		targetEvent = filter.RetentionEvent{ID: req.TargetEvent.Id, Name: req.TargetEvent.Name}
	}
	if req.ReturnEvent != nil {
		returnEvent = filter.RetentionEvent{ID: req.ReturnEvent.Id, Name: req.ReturnEvent.Name}
	}

	result, err := s.analyticsDB.QueryRetention(ctx, &filter.RetentionQuery{
		ProjectID:     req.ProjectId,
		TargetEvent:   targetEvent,
		ReturnEvent:   returnEvent,
		DateFrom:      req.DateFrom,
		DateTo:        req.DateTo,
		Period:        req.Period,
		RetentionType: req.RetentionType,
	})
	if err != nil {
		logger.Error(ctx, "query retention: %v", err)
		return nil, status.Error(codes.Internal, "failed to query retention")
	}

	var rows []*openclick_v1.RetentionCohort
	for _, cohort := range result.Result {
		var values []*openclick_v1.RetentionValue
		for _, v := range cohort.Values {
			values = append(values, &openclick_v1.RetentionValue{
				Count:      v.Count,
				Percentage: v.Percentage,
			})
		}
		rows = append(rows, &openclick_v1.RetentionCohort{
			Date:       cohort.Date.Format("2006-01-02"),
			Label:      cohort.Label,
			CohortSize: cohort.CohortSize,
			Values:     values,
		})
	}
	if rows == nil {
		rows = []*openclick_v1.RetentionCohort{}
	}
	return &openclick_v1.QueryRetentionResponse{Result: rows}, nil
}

func (s *Service) QueryPaths(ctx context.Context, req *openclick_v1.QueryPathsRequest) (*openclick_v1.QueryPathsResponse, error) {
	if err := s.checkAnalyticsAuth(ctx, req.ProjectId, constants.PermAnalyticsRead); err != nil {
		return nil, err
	}

	if s.analyticsDB == nil {
		return &openclick_v1.QueryPathsResponse{Nodes: []*openclick_v1.PathNode{}, Links: []*openclick_v1.PathLink{}}, nil
	}

	result, err := s.analyticsDB.QueryPaths(ctx, &filter.PathsQuery{
		ProjectID:     req.ProjectId,
		DateFrom:      req.DateFrom,
		DateTo:        req.DateTo,
		StartPoint:    req.StartPoint,
		EndPoint:      req.EndPoint,
		PathType:      req.PathType,
		StepLimit:     int(req.StepLimit),
		MinEdgeWeight: int(req.MinEdgeWeight),
	})
	if err != nil {
		logger.Error(ctx, "query paths: %v", err)
		return nil, status.Error(codes.Internal, "failed to query paths")
	}

	var nodes []*openclick_v1.PathNode
	var links []*openclick_v1.PathLink
	for _, n := range result.Nodes {
		nodes = append(nodes, &openclick_v1.PathNode{Id: n.ID, Name: n.Name})
	}
	for _, l := range result.Links {
		links = append(links, &openclick_v1.PathLink{Source: l.Source, Target: l.Target, Value: int64(l.Value)})
	}
	if nodes == nil {
		nodes = []*openclick_v1.PathNode{}
	}
	if links == nil {
		links = []*openclick_v1.PathLink{}
	}
	return &openclick_v1.QueryPathsResponse{Nodes: nodes, Links: links}, nil
}

func (s *Service) QueryEvents(ctx context.Context, req *openclick_v1.QueryEventsRequest) (*openclick_v1.QueryEventsResponse, error) {
	if err := s.checkAnalyticsAuth(ctx, req.ProjectId, constants.PermEventsRead); err != nil {
		return nil, err
	}

	if s.analyticsDB == nil {
		return &openclick_v1.QueryEventsResponse{Results: []*openclick_v1.EventResult{}, Total: 0}, nil
	}

	distinctID := ""
	if req.DistinctId != nil {
		distinctID = *req.DistinctId
	}

	result, err := s.analyticsDB.QueryEvents(ctx, &filter.EventsQuery{
		ProjectID:  req.ProjectId,
		Event:      req.Event,
		DateFrom:   req.DateFrom,
		DateTo:     req.DateTo,
		DistinctID: distinctID,
		Filters:    protoToPropertyFilters(req.Filters),
		Limit:      int(req.Limit),
		Offset:     int(req.Offset),
		OrderBy:    req.OrderBy,
		OrderDir:   req.OrderDir,
	})
	if err != nil {
		logger.Error(ctx, "query events: %v", err)
		return nil, status.Error(codes.Internal, "failed to query events")
	}

	var events []*openclick_v1.EventResult
	for _, e := range result.Results {
		var props structpb.Struct
		if len(e.Properties) > 0 {
			_ = props.UnmarshalJSON([]byte(e.Properties))
		}
		events = append(events, &openclick_v1.EventResult{
			Uuid:       e.UUID,
			Event:      e.Event,
			DistinctId: e.DistinctID,
			Timestamp:  timestamppb.New(e.Timestamp),
			Properties: &props,
		})
	}
	if events == nil {
		events = []*openclick_v1.EventResult{}
	}
	return &openclick_v1.QueryEventsResponse{
		Results: events,
		Total:   int64(result.Total),
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Sessions
// ─────────────────────────────────────────────────────────────────────────────

func (s *Service) ListSessions(ctx context.Context, req *openclick_v1.ListSessionsRequest) (*openclick_v1.ListSessionsResponse, error) {
	if err := s.checkAnalyticsAuth(ctx, req.ProjectId, constants.PermReplayRead); err != nil {
		return nil, err
	}

	if s.analyticsDB == nil {
		return &openclick_v1.ListSessionsResponse{Results: []*openclick_v1.SessionResponse{}, Total: 0}, nil
	}

	df := ""
	if req.DateFrom != nil {
		df = *req.DateFrom
	}
	dt := ""
	if req.DateTo != nil {
		dt = *req.DateTo
	}
	did := ""
	if req.DistinctId != nil {
		did = *req.DistinctId
	}
	search := ""
	if req.Search != nil {
		search = *req.Search
	}
	minDur := 0
	if req.MinDurationMs != nil {
		minDur = int(*req.MinDurationMs)
	}
	limit := 50
	if req.Limit != nil {
		limit = int(*req.Limit)
	}
	offset := 0
	if req.Offset != nil {
		offset = int(*req.Offset)
	}

	f := &filter.SessionFilter{
		ProjectID:     req.ProjectId,
		DateFrom:      df,
		DateTo:        dt,
		DistinctID:    did,
		MinDurationMs: minDur,
		Search:        search,
		Limit:         limit,
		Offset:        offset,
	}

	sessions, total, err := s.analyticsDB.ListSessions(ctx, f)
	if err != nil {
		logger.Error(ctx, "list sessions: %v", err)
		return nil, status.Error(codes.Internal, "failed to list sessions")
	}

	var results []*openclick_v1.SessionResponse
	for _, sess := range sessions {
		results = append(results, daoSessionToResponse(sess))
	}
	if results == nil {
		results = []*openclick_v1.SessionResponse{}
	}
	return &openclick_v1.ListSessionsResponse{Results: results, Total: int64(total)}, nil
}

func (s *Service) GetSession(ctx context.Context, req *openclick_v1.GetSessionRequest) (*openclick_v1.SessionResponse, error) {
	if err := s.checkAnalyticsAuth(ctx, req.ProjectId, constants.PermReplayRead); err != nil {
		return nil, err
	}

	if s.analyticsDB == nil {
		return nil, status.Error(codes.Unavailable, "analytics database not configured")
	}

	sess, err := s.analyticsDB.GetSession(ctx, req.ProjectId, req.SessionId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return daoSessionToResponse(sess), nil
}

func (s *Service) GetSessionChunks(ctx context.Context, req *openclick_v1.GetSessionChunksRequest) (*openclick_v1.GetSessionChunksResponse, error) {
	if err := s.checkAnalyticsAuth(ctx, req.ProjectId, constants.PermReplayRead); err != nil {
		return nil, err
	}

	if s.analyticsDB == nil {
		return nil, status.Error(codes.Unavailable, "analytics database not configured")
	}

	fromChunk := 0
	if req.FromChunk != nil {
		fromChunk = int(*req.FromChunk)
	}

	chunks, total, err := s.analyticsDB.GetSessionChunks(ctx, req.ProjectId, req.SessionId, fromChunk)
	if err != nil {
		logger.Error(ctx, "get session chunks: %v", err)
		return nil, status.Error(codes.Internal, "failed to get session chunks")
	}

	var chunkList []*openclick_v1.SessionChunk
	for _, c := range chunks {
		chunkList = append(chunkList, &openclick_v1.SessionChunk{
			ChunkIndex: int32(c.ChunkIndex),
			Data:       c.Data,
			Timestamp:  timestamppb.New(c.Timestamp),
		})
	}
	if chunkList == nil {
		chunkList = []*openclick_v1.SessionChunk{}
	}
	return &openclick_v1.GetSessionChunksResponse{
		Chunks:      chunkList,
		TotalChunks: int64(total),
	}, nil
}

func (s *Service) DeleteSession(ctx context.Context, req *openclick_v1.DeleteSessionRequest) (*openclick_v1.DeleteSessionResponse, error) {
	if err := s.checkAnalyticsAuth(ctx, req.ProjectId, constants.PermReplayDelete); err != nil {
		return nil, err
	}

	if s.analyticsDB == nil {
		return nil, status.Error(codes.Unavailable, "analytics database not configured")
	}

	if err := s.analyticsDB.DeleteSession(ctx, req.ProjectId, req.SessionId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &openclick_v1.DeleteSessionResponse{}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func (s *Service) checkAnalyticsAuth(ctx context.Context, projectID, perm string) error {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return err
	}
	if !s.hasPermission(ctx, perm) {
		return status.Error(codes.PermissionDenied, "missing permission: "+perm)
	}
	return s.validateMembership(ctx, projectID, userID)
}

func protoToPropertyFilters(pfs []*openclick_v1.PropertyFilter) []filter.PropertyFilter {
	var result []filter.PropertyFilter
	for _, pf := range pfs {
		var val interface{}
		if pf.Value != nil {
			val = pf.Value.AsInterface()
		}
		result = append(result, filter.PropertyFilter{
			Key:      pf.Key,
			Value:    val,
			Operator: pf.Operator,
			Type:     pf.Type,
		})
	}
	return result
}
