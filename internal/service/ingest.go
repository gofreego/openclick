package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/openclick/api/openclick_v1"
	"github.com/gofreego/openclick/internal/models/dao"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ─────────────────────────────────────────────────────────────────────────────
// Ingest handlers — public-facing endpoints authenticated via api_key
// ─────────────────────────────────────────────────────────────────────────────

func (s *Service) CaptureEvent(ctx context.Context, req *openclick_v1.CaptureEventRequest) (*openclick_v1.CaptureEventResponse, error) {
	apiKey := ""
	if req.ApiKey != nil {
		apiKey = *req.ApiKey
	}
	if apiKey == "" {
		apiKey = extractBearerTokenFromCtx(ctx)
	}

	if apiKey == "" {
		return nil, status.Error(codes.Unauthenticated, "api_key is required")
	}
	if req.DistinctId == "" {
		return nil, status.Error(codes.InvalidArgument, "distinct_id is required")
	}
	if req.Event == "" {
		return nil, status.Error(codes.InvalidArgument, "event is required")
	}

	project, err := s.repo.GetProjectByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid api_key")
	}

	ts := time.Now().UTC()
	if req.Timestamp != nil {
		ts = req.Timestamp.AsTime().UTC()
	}

	propsJSON := "{}"
	if req.Properties != nil {
		b, _ := req.Properties.MarshalJSON()
		propsJSON = string(b)
	}

	sessionID := ""
	if req.SessionId != nil {
		sessionID = *req.SessionId
	}

	event := &dao.Event{
		ProjectID:  project.ID,
		UUID:       uuid.New().String(),
		Event:      req.Event,
		DistinctID: req.DistinctId,
		Properties: propsJSON,
		Timestamp:  ts,
		SessionID:  sessionID,
	}

	if s.ingest != nil {
		s.ingest.Push(ctx, event)
	} else {
		logger.Debug(ctx, "ClickHouse not configured, event dropped: %s", req.Event)
	}

	return &openclick_v1.CaptureEventResponse{Status: 1}, nil
}

func (s *Service) BatchCapture(ctx context.Context, req *openclick_v1.BatchCaptureRequest) (*openclick_v1.BatchCaptureResponse, error) {
	apiKey := ""
	if req.ApiKey != nil {
		apiKey = *req.ApiKey
	}
	if apiKey == "" {
		apiKey = extractBearerTokenFromCtx(ctx)
	}

	if apiKey == "" {
		return nil, status.Error(codes.Unauthenticated, "api_key is required")
	}

	if len(req.Batch) > 1000 {
		return nil, status.Error(codes.InvalidArgument, "max batch size is 1000 events")
	}

	project, err := s.repo.GetProjectByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid api_key")
	}

	var events []*dao.Event
	for _, be := range req.Batch {
		if be.DistinctId == "" || be.Event == "" {
			continue
		}
		ts := time.Now().UTC()
		if be.Timestamp != nil {
			ts = be.Timestamp.AsTime().UTC()
		}

		propsJSON := "{}"
		if be.Properties != nil {
			b, _ := be.Properties.MarshalJSON()
			propsJSON = string(b)
		}

		sessionID := ""
		if be.SessionId != nil {
			sessionID = *be.SessionId
		}

		events = append(events, &dao.Event{
			ProjectID:  project.ID,
			UUID:       uuid.New().String(),
			Event:      be.Event,
			DistinctID: be.DistinctId,
			Properties: propsJSON,
			Timestamp:  ts,
			SessionID:  sessionID,
		})
	}

	if s.ingest != nil && len(events) > 0 {
		s.ingest.Push(ctx, events...)
	}

	return &openclick_v1.BatchCaptureResponse{
		Status: 1,
		Count:  int32(len(events)),
	}, nil
}

func (s *Service) Identify(ctx context.Context, req *openclick_v1.IdentifyRequest) (*openclick_v1.IdentifyResponse, error) {
	apiKey := ""
	if req.ApiKey != nil {
		apiKey = *req.ApiKey
	}
	if apiKey == "" {
		apiKey = extractBearerTokenFromCtx(ctx)
	}

	if apiKey == "" {
		return nil, status.Error(codes.Unauthenticated, "api_key is required")
	}
	if req.DistinctId == "" {
		return nil, status.Error(codes.InvalidArgument, "distinct_id is required")
	}

	project, err := s.repo.GetProjectByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid api_key")
	}

	setProps := map[string]interface{}{}
	if req.Properties != nil {
		propsMap := req.Properties.AsMap()
		if set, ok := propsMap["$set"].(map[string]interface{}); ok {
			for k, v := range set {
				setProps[k] = v
			}
		}
		if setOnce, ok := propsMap["$set_once"].(map[string]interface{}); ok {
			for k, v := range setOnce {
				if _, exists := setProps[k]; !exists {
					setProps[k] = v
				}
			}
		}
	}

	propsJSON, _ := json.Marshal(setProps)
	_, err = s.repo.UpsertPerson(ctx, &dao.Person{
		ProjectID:  project.ID,
		DistinctID: req.DistinctId,
		Properties: propsJSON,
	})
	if err != nil {
		logger.Error(ctx, "identify upsert person: %v", err)
	}

	return &openclick_v1.IdentifyResponse{Status: 1}, nil
}

func (s *Service) Alias(ctx context.Context, req *openclick_v1.AliasRequest) (*openclick_v1.AliasResponse, error) {
	apiKey := ""
	if req.ApiKey != nil {
		apiKey = *req.ApiKey
	}
	if apiKey == "" {
		apiKey = extractBearerTokenFromCtx(ctx)
	}

	if apiKey == "" {
		return nil, status.Error(codes.Unauthenticated, "api_key is required")
	}

	_, err := s.repo.GetProjectByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid api_key")
	}

	if s.ingest != nil && req.Alias != "" && req.DistinctId != "" {
		props, _ := json.Marshal(map[string]string{
			"$identified_id": req.Alias,
			"$anon_id":       req.DistinctId,
		})
		s.ingest.Push(ctx, &dao.Event{
			UUID:       uuid.New().String(),
			Event:      "$alias",
			DistinctID: req.DistinctId,
			Properties: string(props),
			Timestamp:  time.Now().UTC(),
		})
	}

	return &openclick_v1.AliasResponse{Status: 1}, nil
}

func (s *Service) IngestReplay(ctx context.Context, req *openclick_v1.IngestReplayRequest) (*openclick_v1.IngestReplayResponse, error) {
	apiKey := ""
	if req.ApiKey != nil {
		apiKey = *req.ApiKey
	}
	if apiKey == "" {
		apiKey = extractBearerTokenFromCtx(ctx)
	}

	if apiKey == "" {
		return nil, status.Error(codes.Unauthenticated, "api_key is required")
	}

	project, err := s.repo.GetProjectByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid api_key")
	}

	if s.analyticsDB == nil {
		return &openclick_v1.IngestReplayResponse{Status: 1}, nil
	}

	chunk := &dao.ReplayChunk{
		ProjectID:  project.ID,
		SessionID:  req.SessionId,
		ChunkIndex: uint16(req.ChunkIndex),
		Data:       req.Data,
		Compressed: true,
		Timestamp:  time.Now().UTC(),
	}
	if err := s.analyticsDB.InsertReplayChunk(ctx, chunk); err != nil {
		logger.Error(ctx, "insert replay chunk: %v", err)
		return nil, status.Error(codes.Internal, "failed to ingest replay chunk")
	}

	return &openclick_v1.IngestReplayResponse{Status: 1}, nil
}
