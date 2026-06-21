package service

import (
	"github.com/gofreego/openclick/api/openclick_v1"
	"github.com/gofreego/openclick/internal/models/dao"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// daoSessionToResponse converts a Session DAO to an API response
func daoSessionToResponse(sess *dao.Session) *openclick_v1.SessionResponse {
	if sess == nil {
		return nil
	}
	return &openclick_v1.SessionResponse{
		ProjectId:    sess.ProjectID,
		SessionId:    sess.SessionID,
		DistinctId:   sess.DistinctID,
		StartTime:    timestamppb.New(sess.StartTime),
		EndTime:      timestamppb.New(sess.EndTime),
		DurationMs:   int64(sess.DurationMs),
		PageCount:    int32(sess.PageCount),
		ClickCount:   int32(sess.ClickCount),
		CountryCode:  sess.CountryCode,
		Browser:      sess.Browser,
		Os:           sess.OS,
		DeviceType:   sess.DeviceType,
		RecordingUrl: sess.RecordingURL,
	}
}
