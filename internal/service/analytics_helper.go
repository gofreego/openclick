package service

import (
	"github.com/gofreego/openclick/internal/models/dao"
)

// daoSessionToResponse converts a Session DAO to an API response
func daoSessionToResponse(sess *dao.Session) map[string]interface{} {
	if sess == nil {
		return nil
	}
	return map[string]interface{}{
		"project_id":    sess.ProjectID,
		"session_id":    sess.SessionID,
		"distinct_id":   sess.DistinctID,
		"start_time":    sess.StartTime,
		"end_time":      sess.EndTime,
		"duration_ms":   sess.DurationMs,
		"page_count":    sess.PageCount,
		"click_count":   sess.ClickCount,
		"country_code":  sess.CountryCode,
		"browser":       sess.Browser,
		"os":            sess.OS,
		"device_type":   sess.DeviceType,
		"recording_url": sess.RecordingURL,
	}
}
