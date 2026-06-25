package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
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

func (s *Service) RegisterDevice(ctx context.Context, req *openclick_v1.RegisterDeviceRequest) (*openclick_v1.RegisterDeviceResponse, error) {
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

	var rawProps map[string]any
	if req.Properties != nil {
		b, _ := req.Properties.MarshalJSON()
		json.Unmarshal(b, &rawProps)
	}

	deviceID, deviceProps, _ := extractDeviceProps(project.ID, rawProps)
	if deviceID == "" {
		return nil, status.Error(codes.InvalidArgument, "no device properties provided")
	}

	s.upsertDeviceCached(ctx, project.ID, deviceID, deviceProps)

	return &openclick_v1.RegisterDeviceResponse{DeviceId: deviceID}, nil
}

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

	deviceID := ""
	if req.DeviceId != nil {
		deviceID = *req.DeviceId
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
		DeviceID:   deviceID,
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

		deviceID := ""
		if be.DeviceId != nil {
			deviceID = *be.DeviceId
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
			DeviceID:   deviceID,
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

// ─────────────────────────────────────────────────────────────────────────────
// Device helpers
// ─────────────────────────────────────────────────────────────────────────────

// devicePropKeys is the set of event property keys that belong to a device.
var devicePropKeys = map[string]bool{
	"$browser": true, "$browser_version": true,
	"$device_type": true,
	"$os": true, "$os_version": true,
	"$lib": true, "$lib_version": true,
	"$screen_height": true, "$screen_width": true,
	"$viewport_height": true, "$viewport_width": true,
	"$referrer": true, "$referring_domain": true,
	"$user_agent": true,
	"$device_id": true,
}

// stableFingerprintKeys are the subset used to generate a deterministic device ID.
var stableFingerprintKeys = []string{
	"$browser", "$device_type", "$os", "$screen_height", "$screen_width", "$lib",
}

// extractDeviceProps splits a raw property map into (deviceID, deviceProps, remainingProps).
// If $device_id is present it is used directly; otherwise a SHA-256 fingerprint of
// stable hardware/browser properties is generated.
// Returns empty deviceID if no device properties are found.
func extractDeviceProps(projectID string, props map[string]interface{}) (string, map[string]interface{}, map[string]interface{}) {
	deviceProps := make(map[string]interface{})
	remaining := make(map[string]interface{})

	for k, v := range props {
		if devicePropKeys[k] {
			deviceProps[k] = v
		} else {
			remaining[k] = v
		}
	}

	if len(deviceProps) == 0 {
		return "", deviceProps, remaining
	}

	// Use explicit $device_id if the SDK provided one.
	if id, ok := deviceProps["$device_id"].(string); ok && id != "" {
		return id, deviceProps, remaining
	}

	// Generate a deterministic fingerprint from stable properties.
	parts := []string{projectID}
	for _, k := range stableFingerprintKeys {
		if v, ok := deviceProps[k]; ok {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
	}
	sort.Strings(parts[1:]) // keep projectID first, sort the rest for stability
	h := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(h[:16]), deviceProps, remaining
}

// upsertDeviceCached upserts a device into PostgreSQL, skipping the DB call if
// this (projectID, deviceID) pair was already written during this process lifetime.
func (s *Service) upsertDeviceCached(ctx context.Context, projectID, deviceID string, props map[string]interface{}) {
	cacheKey := projectID + ":" + deviceID
	if _, seen := s.deviceCache.Load(cacheKey); seen {
		return
	}
	propsJSON, _ := json.Marshal(props)
	if _, err := s.repo.UpsertDevice(ctx, &dao.Device{
		ID:         deviceID,
		ProjectID:  projectID,
		Properties: propsJSON,
	}); err != nil {
		logger.Error(ctx, "upsert device %s: %v", deviceID, err)
		return
	}
	s.deviceCache.Store(cacheKey, struct{}{})
}
