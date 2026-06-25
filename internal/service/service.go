package service

import (
	"context"
	"sync"

	"github.com/gofreego/openclick/api/openclick_v1"
	"github.com/gofreego/openclick/internal/models/dao"
	"github.com/gofreego/openclick/internal/models/filter"
	"github.com/gofreego/openclick/internal/repository/clickhouse"
)

// Config holds service-level configuration
type Config struct{}

// Repository defines all data operations for the service layer.
// The concrete implementation can be PostgreSQL (for metadata) with ClickHouse (for analytics).
type Repository interface {
	Ping(ctx context.Context) error

	// Projects
	CreateProject(ctx context.Context, p *dao.Project) (*dao.Project, error)
	GetProjectByID(ctx context.Context, id string) (*dao.Project, error)
	GetProjectByAPIKey(ctx context.Context, apiKey string) (*dao.Project, error)
	GetProjectBySecretKey(ctx context.Context, secretKey string) (*dao.Project, error)
	ListProjectsByUserID(ctx context.Context, userID string) ([]*dao.Project, error)
	UpdateProject(ctx context.Context, p *dao.Project) (*dao.Project, error)
	DeleteProject(ctx context.Context, id string) error

	// Project Members
	AddProjectMember(ctx context.Context, m *dao.ProjectMember) (*dao.ProjectMember, error)
	RemoveProjectMember(ctx context.Context, projectID, userID string) error
	GetProjectMembers(ctx context.Context, projectID string) ([]*dao.ProjectMember, error)
	IsProjectMember(ctx context.Context, projectID, userID string) (bool, error)

	// Devices
	UpsertDevice(ctx context.Context, d *dao.Device) (*dao.Device, error)
	ListDevices(ctx context.Context, f *filter.DeviceFilter) ([]*dao.Device, int, error)
	GetDevice(ctx context.Context, projectID, deviceID string) (*dao.Device, error)
	GetDeviceStats(ctx context.Context, projectID string) (browsers, osList, deviceTypes, libs []dao.StatItem, err error)

	// Persons
	UpsertPerson(ctx context.Context, p *dao.Person) (*dao.Person, error)
	GetPerson(ctx context.Context, projectID, distinctID string) (*dao.Person, error)
	ListPersons(ctx context.Context, f *filter.PersonFilter) ([]*dao.Person, int, error)
	DeletePerson(ctx context.Context, projectID, distinctID string) error

	// Cohorts
	CreateCohort(ctx context.Context, c *dao.Cohort) (*dao.Cohort, error)
	ListCohorts(ctx context.Context, f *filter.CohortFilter) ([]*dao.Cohort, error)
	DeleteCohort(ctx context.Context, projectID, cohortID string) error

	// Feature Flags
	CreateFeatureFlag(ctx context.Context, f *dao.FeatureFlag) (*dao.FeatureFlag, error)
	GetFeatureFlagByID(ctx context.Context, projectID, flagID string) (*dao.FeatureFlag, error)
	ListFeatureFlags(ctx context.Context, f *filter.FeatureFlagFilter) ([]*dao.FeatureFlag, error)
	ListActiveFeatureFlags(ctx context.Context, projectID string) ([]*dao.FeatureFlag, error)
	UpdateFeatureFlag(ctx context.Context, f *dao.FeatureFlag) (*dao.FeatureFlag, error)
	DeleteFeatureFlag(ctx context.Context, projectID, flagID string) error

	// Dashboards
	CreateDashboard(ctx context.Context, d *dao.Dashboard) (*dao.Dashboard, error)
	GetDashboard(ctx context.Context, projectID, dashboardID string) (*dao.Dashboard, []*dao.DashboardItem, error)
	ListDashboards(ctx context.Context, f *filter.DashboardFilter) ([]*dao.Dashboard, []int, error)
	DeleteDashboard(ctx context.Context, projectID, dashboardID string) error

	// Dashboard Items
	CreateDashboardItem(ctx context.Context, item *dao.DashboardItem) (*dao.DashboardItem, error)
	UpdateDashboardItem(ctx context.Context, item *dao.DashboardItem) (*dao.DashboardItem, error)
	DeleteDashboardItem(ctx context.Context, dashboardID, itemID string) error
}

// AnalyticsRepository defines ClickHouse-specific data operations
type AnalyticsRepository interface {
	Ping(ctx context.Context) error

	// Event Ingest
	InsertEvents(ctx context.Context, events []*dao.Event) error
	UpsertSession(ctx context.Context, s *dao.Session) error
	InsertReplayChunk(ctx context.Context, chunk *dao.ReplayChunk) error

	// Sessions
	ListSessions(ctx context.Context, f *filter.SessionFilter) ([]*dao.Session, int, error)
	GetSession(ctx context.Context, projectID, sessionID string) (*dao.Session, error)
	GetSessionChunks(ctx context.Context, projectID, sessionID string, fromChunk int) ([]*dao.ReplayChunk, int, error)
	DeleteSession(ctx context.Context, projectID, sessionID string) error

	// Analytics Queries
	QueryTrends(ctx context.Context, q *filter.TrendsQuery) (*clickhouse.TrendsResult, error)
	QueryFunnel(ctx context.Context, q *filter.FunnelQuery) (*clickhouse.FunnelResult, error)
	QueryRetention(ctx context.Context, q *filter.RetentionQuery) (*clickhouse.RetentionResult, error)
	QueryPaths(ctx context.Context, q *filter.PathsQuery) (*clickhouse.PathsResult, error)
	QueryEvents(ctx context.Context, q *filter.EventsQuery) (*clickhouse.EventsResult, error)
	ListEventNames(ctx context.Context, projectID string) ([]string, error)
}

// Service holds all dependencies for the business logic layer
type Service struct {
	repo        Repository
	analyticsDB AnalyticsRepository // may be nil if ClickHouse is not configured
	ingest      *IngestBuffer
	deviceCache sync.Map // tracks "projectID:deviceID" already upserted this session
	openclick_v1.UnimplementedBaseServiceServer
}

// NewService creates a new Service with the given repositories
func NewService(ctx context.Context, cfg *Config, repo Repository, analyticsDB AnalyticsRepository) *Service {
	svc := &Service{
		repo:        repo,
		analyticsDB: analyticsDB,
	}
	// Only start the ingest buffer if ClickHouse is available
	if analyticsDB != nil {
		svc.ingest = NewIngestBuffer(ctx, analyticsDB, 1000)
	}
	return svc
}

