package service

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/openclick/internal/models/dao"
)

// ─────────────────────────────────────────────────────────────────────────────
// Ingest handlers — public-facing endpoints authenticated via api_key
// ─────────────────────────────────────────────────────────────────────────────

// CaptureEvent handles POST /e/
func (s *Service) CaptureEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body struct {
		APIKey     string                 `json:"api_key"`
		Event      string                 `json:"event"`
		DistinctID string                 `json:"distinct_id"`
		Timestamp  *time.Time             `json:"timestamp"`
		Properties map[string]interface{} `json:"properties"`
		SessionID  string                 `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	// Resolve api_key from body or Authorization header
	apiKey := body.APIKey
	if apiKey == "" {
		apiKey = extractBearerToken(r)
	}
	if apiKey == "" {
		writeError(w, http.StatusUnauthorized, "api_key is required", "UNAUTHORIZED")
		return
	}
	if body.DistinctID == "" {
		writeError(w, http.StatusBadRequest, "distinct_id is required", "VALIDATION_ERROR")
		return
	}
	if body.Event == "" {
		writeError(w, http.StatusBadRequest, "event is required", "VALIDATION_ERROR")
		return
	}

	project, err := s.repo.GetProjectByAPIKey(ctx, apiKey)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid api_key", "UNAUTHORIZED")
		return
	}

	ts := time.Now().UTC()
	if body.Timestamp != nil {
		ts = body.Timestamp.UTC()
	}

	propsJSON := "{}"
	if body.Properties != nil {
		b, _ := json.Marshal(body.Properties)
		propsJSON = string(b)
	}

	event := &dao.Event{
		ProjectID:  project.ID,
		UUID:       uuid.New().String(),
		Event:      body.Event,
		DistinctID: body.DistinctID,
		Properties: propsJSON,
		Timestamp:  ts,
		SessionID:  body.SessionID,
	}

	if s.ingest != nil {
		s.ingest.Push(ctx, event)
	} else {
		logger.Debug(ctx, "ClickHouse not configured, event dropped: %s", body.Event)
	}

	writeJSON(w, http.StatusOK, map[string]int{"status": 1})
}

// BatchCapture handles POST /batch/
func (s *Service) BatchCapture(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	type batchEvent struct {
		Event      string                 `json:"event"`
		DistinctID string                 `json:"distinct_id"`
		Timestamp  *time.Time             `json:"timestamp"`
		Properties map[string]interface{} `json:"properties"`
		SessionID  string                 `json:"session_id"`
	}
	var body struct {
		APIKey string       `json:"api_key"`
		Batch  []batchEvent `json:"batch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	apiKey := body.APIKey
	if apiKey == "" {
		apiKey = extractBearerToken(r)
	}
	if apiKey == "" {
		writeError(w, http.StatusUnauthorized, "api_key is required", "UNAUTHORIZED")
		return
	}

	if len(body.Batch) > 1000 {
		writeError(w, http.StatusBadRequest, "max batch size is 1000 events", "VALIDATION_ERROR")
		return
	}

	project, err := s.repo.GetProjectByAPIKey(ctx, apiKey)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid api_key", "UNAUTHORIZED")
		return
	}

	var events []*dao.Event
	for _, be := range body.Batch {
		if be.DistinctID == "" || be.Event == "" {
			continue
		}
		ts := time.Now().UTC()
		if be.Timestamp != nil {
			ts = be.Timestamp.UTC()
		}
		propsJSON := "{}"
		if be.Properties != nil {
			b, _ := json.Marshal(be.Properties)
			propsJSON = string(b)
		}
		events = append(events, &dao.Event{
			ProjectID:  project.ID,
			UUID:       uuid.New().String(),
			Event:      be.Event,
			DistinctID: be.DistinctID,
			Properties: propsJSON,
			Timestamp:  ts,
			SessionID:  be.SessionID,
		})
	}

	if s.ingest != nil && len(events) > 0 {
		s.ingest.Push(ctx, events...)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": 1,
		"count":  len(events),
	})
}

// Identify handles POST /identify/
func (s *Service) Identify(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body struct {
		APIKey     string                 `json:"api_key"`
		DistinctID string                 `json:"distinct_id"`
		Properties map[string]interface{} `json:"properties"` // contains $set and $set_once
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	apiKey := body.APIKey
	if apiKey == "" {
		apiKey = extractBearerToken(r)
	}
	if apiKey == "" {
		writeError(w, http.StatusUnauthorized, "api_key is required", "UNAUTHORIZED")
		return
	}
	if body.DistinctID == "" {
		writeError(w, http.StatusBadRequest, "distinct_id is required", "VALIDATION_ERROR")
		return
	}

	project, err := s.repo.GetProjectByAPIKey(ctx, apiKey)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid api_key", "UNAUTHORIZED")
		return
	}

	// Extract $set properties and merge into person
	setProps := map[string]interface{}{}
	if body.Properties != nil {
		if set, ok := body.Properties["$set"].(map[string]interface{}); ok {
			for k, v := range set {
				setProps[k] = v
			}
		}
		// $set_once only sets if not already present — we simplify and merge all
		if setOnce, ok := body.Properties["$set_once"].(map[string]interface{}); ok {
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
		DistinctID: body.DistinctID,
		Properties: propsJSON,
	})
	if err != nil {
		logger.Error(ctx, "identify upsert person: %v", err)
	}

	writeJSON(w, http.StatusOK, map[string]int{"status": 1})
}

// Alias handles POST /alias/
func (s *Service) Alias(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body struct {
		APIKey     string `json:"api_key"`
		Alias      string `json:"alias"`       // the new identified user
		DistinctID string `json:"distinct_id"` // the anonymous ID
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	apiKey := body.APIKey
	if apiKey == "" {
		apiKey = extractBearerToken(r)
	}
	if apiKey == "" {
		writeError(w, http.StatusUnauthorized, "api_key is required", "UNAUTHORIZED")
		return
	}

	_, err := s.repo.GetProjectByAPIKey(ctx, apiKey)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid api_key", "UNAUTHORIZED")
		return
	}

	// Alias is recorded as a special $identify event
	if s.ingest != nil && body.Alias != "" && body.DistinctID != "" {
		props, _ := json.Marshal(map[string]string{
			"$identified_id": body.Alias,
			"$anon_id":       body.DistinctID,
		})
		s.ingest.Push(ctx, &dao.Event{
			UUID:       uuid.New().String(),
			Event:      "$alias",
			DistinctID: body.DistinctID,
			Properties: string(props),
			Timestamp:  time.Now().UTC(),
		})
	}

	writeJSON(w, http.StatusOK, map[string]int{"status": 1})
}

// IngestReplay handles POST /replay/
func (s *Service) IngestReplay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body struct {
		APIKey     string            `json:"api_key"`
		SessionID  string            `json:"session_id"`
		DistinctID string            `json:"distinct_id"`
		ChunkIndex uint16            `json:"chunk_index"`
		Data       string            `json:"data"` // base64-encoded compressed rrweb JSON
		Metadata   map[string]string `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	apiKey := body.APIKey
	if apiKey == "" {
		apiKey = extractBearerToken(r)
	}
	if apiKey == "" {
		writeError(w, http.StatusUnauthorized, "api_key is required", "UNAUTHORIZED")
		return
	}

	project, err := s.repo.GetProjectByAPIKey(ctx, apiKey)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid api_key", "UNAUTHORIZED")
		return
	}

	if s.analyticsDB == nil {
		writeJSON(w, http.StatusOK, map[string]int{"status": 1})
		return
	}

	chunk := &dao.ReplayChunk{
		ProjectID:  project.ID,
		SessionID:  body.SessionID,
		ChunkIndex: body.ChunkIndex,
		Data:       body.Data,
		Compressed: true,
		Timestamp:  time.Now().UTC(),
	}
	if err := s.analyticsDB.InsertReplayChunk(ctx, chunk); err != nil {
		logger.Error(ctx, "insert replay chunk: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to ingest replay chunk", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"status": 1})
}

// extractBearerToken extracts the token from "Authorization: Bearer <token>"
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}
