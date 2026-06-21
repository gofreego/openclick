package memory

import (
	"context"
	"errors"

	"github.com/gofreego/openclick/internal/models/dao"
	"github.com/gofreego/openclick/internal/models/filter"
)

var errNotImpl = errors.New("not implemented in memory repository")

type Config struct{}

type MemoryRepository struct {
	cfg *Config
}

func NewRepository(ctx context.Context, cfg *Config) (*MemoryRepository, error) {
	return &MemoryRepository{cfg: cfg}, nil
}

func (r *MemoryRepository) Ping(ctx context.Context) error { return nil }

func (r *MemoryRepository) CreateProject(ctx context.Context, p *dao.Project) (*dao.Project, error) { return nil, errNotImpl }
func (r *MemoryRepository) GetProjectByID(ctx context.Context, id string) (*dao.Project, error) { return nil, errNotImpl }
func (r *MemoryRepository) GetProjectByAPIKey(ctx context.Context, apiKey string) (*dao.Project, error) { return nil, errNotImpl }
func (r *MemoryRepository) GetProjectBySecretKey(ctx context.Context, secretKey string) (*dao.Project, error) { return nil, errNotImpl }
func (r *MemoryRepository) ListProjectsByUserID(ctx context.Context, userID string) ([]*dao.Project, error) { return nil, errNotImpl }
func (r *MemoryRepository) UpdateProject(ctx context.Context, p *dao.Project) (*dao.Project, error) { return nil, errNotImpl }
func (r *MemoryRepository) DeleteProject(ctx context.Context, id string) error { return errNotImpl }

func (r *MemoryRepository) AddProjectMember(ctx context.Context, m *dao.ProjectMember) (*dao.ProjectMember, error) { return nil, errNotImpl }
func (r *MemoryRepository) RemoveProjectMember(ctx context.Context, projectID, userID string) error { return errNotImpl }
func (r *MemoryRepository) GetProjectMembers(ctx context.Context, projectID string) ([]*dao.ProjectMember, error) { return nil, errNotImpl }
func (r *MemoryRepository) IsProjectMember(ctx context.Context, projectID, userID string) (bool, error) { return false, errNotImpl }

func (r *MemoryRepository) UpsertPerson(ctx context.Context, p *dao.Person) (*dao.Person, error) { return nil, errNotImpl }
func (r *MemoryRepository) GetPerson(ctx context.Context, projectID, distinctID string) (*dao.Person, error) { return nil, errNotImpl }
func (r *MemoryRepository) ListPersons(ctx context.Context, f *filter.PersonFilter) ([]*dao.Person, int, error) { return nil, 0, errNotImpl }
func (r *MemoryRepository) DeletePerson(ctx context.Context, projectID, distinctID string) error { return errNotImpl }

func (r *MemoryRepository) CreateCohort(ctx context.Context, c *dao.Cohort) (*dao.Cohort, error) { return nil, errNotImpl }
func (r *MemoryRepository) ListCohorts(ctx context.Context, f *filter.CohortFilter) ([]*dao.Cohort, error) { return nil, errNotImpl }
func (r *MemoryRepository) DeleteCohort(ctx context.Context, projectID, cohortID string) error { return errNotImpl }

func (r *MemoryRepository) CreateFeatureFlag(ctx context.Context, f *dao.FeatureFlag) (*dao.FeatureFlag, error) { return nil, errNotImpl }
func (r *MemoryRepository) GetFeatureFlagByID(ctx context.Context, projectID, flagID string) (*dao.FeatureFlag, error) { return nil, errNotImpl }
func (r *MemoryRepository) ListFeatureFlags(ctx context.Context, f *filter.FeatureFlagFilter) ([]*dao.FeatureFlag, error) { return nil, errNotImpl }
func (r *MemoryRepository) ListActiveFeatureFlags(ctx context.Context, projectID string) ([]*dao.FeatureFlag, error) { return nil, errNotImpl }
func (r *MemoryRepository) UpdateFeatureFlag(ctx context.Context, f *dao.FeatureFlag) (*dao.FeatureFlag, error) { return nil, errNotImpl }
func (r *MemoryRepository) DeleteFeatureFlag(ctx context.Context, projectID, flagID string) error { return errNotImpl }

func (r *MemoryRepository) CreateDashboard(ctx context.Context, d *dao.Dashboard) (*dao.Dashboard, error) { return nil, errNotImpl }
func (r *MemoryRepository) GetDashboard(ctx context.Context, projectID, dashboardID string) (*dao.Dashboard, []*dao.DashboardItem, error) { return nil, nil, errNotImpl }
func (r *MemoryRepository) ListDashboards(ctx context.Context, f *filter.DashboardFilter) ([]*dao.Dashboard, []int, error) { return nil, nil, errNotImpl }
func (r *MemoryRepository) DeleteDashboard(ctx context.Context, projectID, dashboardID string) error { return errNotImpl }

func (r *MemoryRepository) CreateDashboardItem(ctx context.Context, item *dao.DashboardItem) (*dao.DashboardItem, error) { return nil, errNotImpl }
func (r *MemoryRepository) UpdateDashboardItem(ctx context.Context, item *dao.DashboardItem) (*dao.DashboardItem, error) { return nil, errNotImpl }
func (r *MemoryRepository) DeleteDashboardItem(ctx context.Context, dashboardID, itemID string) error { return errNotImpl }
