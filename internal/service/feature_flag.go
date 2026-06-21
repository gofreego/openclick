package service

import (
	"github.com/gofreego/openclick/internal/constants"

	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/openclick/api/openclick_v1"
	"github.com/gofreego/openclick/internal/models/dao"
	"github.com/gofreego/openclick/internal/models/filter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ─────────────────────────────────────────────────────────────────────────────
// Feature Flags
// ─────────────────────────────────────────────────────────────────────────────

func (s *Service) ListFeatureFlags(ctx context.Context, req *openclick_v1.ListFeatureFlagsRequest) (*openclick_v1.ListFeatureFlagsResponse, error) {
	if err := s.checkFlagAuth(ctx, req.ProjectId, constants.PermFlagsRead); err != nil {
		return nil, err
	}

	flags, err := s.repo.ListFeatureFlags(ctx, &filter.FeatureFlagFilter{ProjectID: req.ProjectId})
	if err != nil {
		logger.Error(ctx, "list feature flags: %v", err)
		return nil, status.Error(codes.Internal, "failed to list feature flags")
	}

	var results []*openclick_v1.FeatureFlagResponse
	for _, f := range flags {
		results = append(results, flagToResponse(f))
	}
	if results == nil {
		results = []*openclick_v1.FeatureFlagResponse{}
	}

	return &openclick_v1.ListFeatureFlagsResponse{Results: results}, nil
}

func (s *Service) CreateFeatureFlag(ctx context.Context, req *openclick_v1.CreateFeatureFlagRequest) (*openclick_v1.FeatureFlagResponse, error) {
	if err := s.checkFlagAuth(ctx, req.ProjectId, constants.PermFlagsWrite); err != nil {
		return nil, err
	}

	if req.Key == "" || req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "key and name are required")
	}

	filtersBytes, _ := req.Filters.MarshalJSON()
	if len(filtersBytes) == 0 || string(filtersBytes) == "null" {
		filtersBytes = []byte("{}")
	}

	f := &dao.FeatureFlag{
		ProjectID:  req.ProjectId,
		Key:        req.Key,
		Name:       req.Name,
		Active:     req.Active,
		RolloutPct: int16(req.RolloutPct),
		Filters:    filtersBytes,
	}
	created, err := s.repo.CreateFeatureFlag(ctx, f)
	if err != nil {
		logger.Error(ctx, "create feature flag: %v", err)
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	return flagToResponse(created), nil
}

func (s *Service) UpdateFeatureFlag(ctx context.Context, req *openclick_v1.UpdateFeatureFlagRequest) (*openclick_v1.FeatureFlagResponse, error) {
	if err := s.checkFlagAuth(ctx, req.ProjectId, constants.PermFlagsWrite); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetFeatureFlagByID(ctx, req.ProjectId, req.FlagId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Active != nil {
		existing.Active = *req.Active
	}
	if req.RolloutPct != nil {
		existing.RolloutPct = int16(*req.RolloutPct)
	}
	if req.Filters != nil {
		filtersBytes, _ := req.Filters.MarshalJSON()
		existing.Filters = filtersBytes
	}

	updated, err := s.repo.UpdateFeatureFlag(ctx, existing)
	if err != nil {
		logger.Error(ctx, "update feature flag: %v", err)
		return nil, status.Error(codes.Internal, "failed to update feature flag")
	}
	return flagToResponse(updated), nil
}

func (s *Service) DeleteFeatureFlag(ctx context.Context, req *openclick_v1.DeleteFeatureFlagRequest) (*openclick_v1.DeleteFeatureFlagResponse, error) {
	if err := s.checkFlagAuth(ctx, req.ProjectId, constants.PermFlagsDelete); err != nil {
		return nil, err
	}

	if err := s.repo.DeleteFeatureFlag(ctx, req.ProjectId, req.FlagId); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &openclick_v1.DeleteFeatureFlagResponse{}, nil
}

func (s *Service) EvaluateFlags(ctx context.Context, req *openclick_v1.EvaluateFlagsRequest) (*openclick_v1.EvaluateFlagsResponse, error) {
	secretKey := extractBearerTokenFromCtx(ctx)
	if secretKey == "" {
		return nil, status.Error(codes.Unauthenticated, "Bearer secret_key is required")
	}

	project, err := s.repo.GetProjectBySecretKey(ctx, secretKey)
	if err != nil || project.ID != req.ProjectId {
		return nil, status.Error(codes.Unauthenticated, "invalid secret_key")
	}

	if req.DistinctId == "" {
		return nil, status.Error(codes.InvalidArgument, "distinct_id is required")
	}

	flags, err := s.repo.ListActiveFeatureFlags(ctx, req.ProjectId)
	if err != nil {
		logger.Error(ctx, "evaluate flags: %v", err)
		return nil, status.Error(codes.Internal, "failed to evaluate flags")
	}

	var personProps map[string]interface{}
	if req.PersonProperties != nil {
		personProps = req.PersonProperties.AsMap()
	}

	results := evaluateFlags(flags, req.DistinctId, personProps)
	return &openclick_v1.EvaluateFlagsResponse{FeatureFlags: results}, nil
}

func (s *Service) Decide(ctx context.Context, req *openclick_v1.DecideRequest) (*openclick_v1.DecideResponse, error) {
	apiKey := ""
	if req.ApiKey != nil {
		apiKey = *req.ApiKey
	}
	if apiKey == "" {
		apiKey = extractBearerTokenFromCtx(ctx)
	}

	if apiKey == "" || req.DistinctId == "" {
		return nil, status.Error(codes.InvalidArgument, "api_key and distinct_id are required")
	}

	project, err := s.repo.GetProjectByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid api_key")
	}

	flags, err := s.repo.ListActiveFeatureFlags(ctx, project.ID)
	if err != nil {
		logger.Error(ctx, "decide: %v", err)
		return nil, status.Error(codes.Internal, "failed to evaluate flags")
	}

	results := evaluateFlags(flags, req.DistinctId, nil)
	return &openclick_v1.DecideResponse{FeatureFlags: results}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Flag evaluation helpers
// ─────────────────────────────────────────────────────────────────────────────

func (s *Service) checkFlagAuth(ctx context.Context, projectID, perm string) error {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return err
	}
	if !s.hasPermission(ctx, perm) {
		return status.Error(codes.PermissionDenied, "missing permission: "+perm)
	}
	return s.validateMembership(ctx, projectID, userID)
}

func extractBearerTokenFromCtx(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return ""
	}
	authHeader := vals[0]
	if len(authHeader) > 7 && strings.EqualFold(authHeader[:7], "bearer ") {
		return authHeader[7:]
	}
	return ""
}

// evaluateFlags determines which flags are enabled for a given user
func evaluateFlags(flags []*dao.FeatureFlag, distinctID string, personProps map[string]interface{}) map[string]bool {
	results := make(map[string]bool, len(flags))
	for _, f := range flags {
		results[f.Key] = isFlagEnabled(f, distinctID, personProps)
	}
	return results
}

// isFlagEnabled determines if a specific flag is enabled for a user
// Uses deterministic hashing on distinctID to split traffic consistently
func isFlagEnabled(f *dao.FeatureFlag, distinctID string, personProps map[string]interface{}) bool {
	if !f.Active {
		return false
	}

	// Check property filter conditions
	var filterCfg struct {
		Groups []struct {
			Properties []struct {
				Key      string      `json:"key"`
				Value    interface{} `json:"value"`
				Operator string      `json:"operator"`
				Type     string      `json:"type"`
			} `json:"properties"`
			RolloutPercentage int `json:"rollout_percentage"`
		} `json:"groups"`
	}
	if len(f.Filters) > 0 && string(f.Filters) != "{}" {
		json.Unmarshal(f.Filters, &filterCfg)
		if len(filterCfg.Groups) > 0 && personProps != nil {
			matchedAny := false
			for _, group := range filterCfg.Groups {
				if matchesGroup(group.Properties, personProps) {
					matchedAny = true
					break
				}
			}
			if !matchedAny {
				return false
			}
		}
	}

	// Rollout percentage check using deterministic hash
	if f.RolloutPct >= 100 {
		return true
	}
	if f.RolloutPct <= 0 {
		return false
	}
	hash := sha256.Sum256([]byte(f.Key + "." + distinctID))
	bucket := int(hash[0]) % 100
	return bucket < int(f.RolloutPct)
}

// matchesGroup checks if personProps satisfies all property conditions in a group
func matchesGroup(conditions []struct {
	Key      string      `json:"key"`
	Value    interface{} `json:"value"`
	Operator string      `json:"operator"`
	Type     string      `json:"type"`
}, personProps map[string]interface{}) bool {
	for _, cond := range conditions {
		val, exists := personProps[cond.Key]
		if !exists {
			return false
		}
		switch cond.Operator {
		case "exact":
			if fmt.Sprintf("%v", val) != fmt.Sprintf("%v", cond.Value) {
				return false
			}
		case "contains":
			if !strings.Contains(strings.ToLower(fmt.Sprintf("%v", val)), strings.ToLower(fmt.Sprintf("%v", cond.Value))) {
				return false
			}
		}
	}
	return true
}

// flagToResponse converts a FeatureFlag DAO to an API response
func flagToResponse(f *dao.FeatureFlag) *openclick_v1.FeatureFlagResponse {
	var filters structpb.Struct
	if len(f.Filters) > 0 {
		_ = filters.UnmarshalJSON(f.Filters)
	}

	return &openclick_v1.FeatureFlagResponse{
		Id:         f.ID,
		Key:        f.Key,
		Name:       f.Name,
		Active:     f.Active,
		RolloutPct: int32(f.RolloutPct),
		Filters:    &filters,
		CreatedAt:  timestamppb.New(f.CreatedAt),
	}
}
