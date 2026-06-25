package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	sqlutils "github.com/gofreego/goutils/databases/connections/sql"
	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/openclick/internal/models/dao"
	"github.com/gofreego/openclick/internal/models/filter"
)

// Repository implements the service.Repository interface using PostgreSQL
type Repository struct {
	connManager sqlutils.DBManager
}

// NewRepository creates a new PostgreSQL repository instance
func NewRepository(ctx context.Context, cfg *sqlutils.Config) (*Repository, error) {
	connManager, err := sqlutils.NewDBManager(cfg)
	if err != nil {
		return nil, err
	}
	return &Repository{connManager: connManager}, nil
}

// Ping checks the database connection
func (r *Repository) Ping(ctx context.Context) error {
	return r.connManager.Primary().Ping()
}

// ─────────────────────────────────────────────────────────────────────────────
// Projects
// ─────────────────────────────────────────────────────────────────────────────

// CreateProject inserts a new project and returns it with generated fields
func (r *Repository) CreateProject(ctx context.Context, p *dao.Project) (*dao.Project, error) {
	query := `
		INSERT INTO projects (name, api_key, secret_key, timezone)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := r.connManager.Primary().QueryRowContext(ctx, query,
		p.Name, p.APIKey, p.SecretKey, p.Timezone,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	return p, nil
}

// GetProjectByID fetches a single project
func (r *Repository) GetProjectByID(ctx context.Context, id string) (*dao.Project, error) {
	query := `
		SELECT id, name, api_key, secret_key, timezone, created_at, updated_at
		FROM projects WHERE id = $1
	`
	var p dao.Project
	err := r.connManager.Primary().QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.APIKey, &p.SecretKey, &p.Timezone, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("project %s not found", id)
		}
		return nil, fmt.Errorf("get project: %w", err)
	}
	return &p, nil
}

// GetProjectByAPIKey fetches a project by its public api_key (used for event ingest auth)
func (r *Repository) GetProjectByAPIKey(ctx context.Context, apiKey string) (*dao.Project, error) {
	query := `SELECT id, name, api_key, secret_key, timezone, created_at, updated_at FROM projects WHERE api_key = $1`
	var p dao.Project
	err := r.connManager.Primary().QueryRowContext(ctx, query, apiKey).Scan(
		&p.ID, &p.Name, &p.APIKey, &p.SecretKey, &p.Timezone, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invalid api_key")
		}
		return nil, fmt.Errorf("get project by api_key: %w", err)
	}
	return &p, nil
}

// GetProjectBySecretKey fetches a project by its secret_key (used for server-side auth)
func (r *Repository) GetProjectBySecretKey(ctx context.Context, secretKey string) (*dao.Project, error) {
	query := `SELECT id, name, api_key, secret_key, timezone, created_at, updated_at FROM projects WHERE secret_key = $1`
	var p dao.Project
	err := r.connManager.Primary().QueryRowContext(ctx, query, secretKey).Scan(
		&p.ID, &p.Name, &p.APIKey, &p.SecretKey, &p.Timezone, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invalid secret_key")
		}
		return nil, fmt.Errorf("get project by secret_key: %w", err)
	}
	return &p, nil
}

// ListProjectsByUserID returns all projects where the given user is a member
func (r *Repository) ListProjectsByUserID(ctx context.Context, userID string) ([]*dao.Project, error) {
	query := `
		SELECT p.id, p.name, p.api_key, p.secret_key, p.timezone, p.created_at, p.updated_at
		FROM projects p
		INNER JOIN project_members pm ON p.id = pm.project_id
		WHERE pm.user_id = $1
		ORDER BY p.created_at DESC
	`
	rows, err := r.connManager.Primary().QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()
	var projects []*dao.Project
	for rows.Next() {
		var p dao.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.APIKey, &p.SecretKey, &p.Timezone, &p.CreatedAt, &p.UpdatedAt); err != nil {
			logger.Error(ctx, "scan project row: %v", err)
			continue
		}
		projects = append(projects, &p)
	}
	return projects, rows.Err()
}

// UpdateProject updates name and/or timezone
func (r *Repository) UpdateProject(ctx context.Context, p *dao.Project) (*dao.Project, error) {
	query := `
		UPDATE projects SET name = $1, timezone = $2, updated_at = now()
		WHERE id = $3
		RETURNING updated_at
	`
	err := r.connManager.Primary().QueryRowContext(ctx, query, p.Name, p.Timezone, p.ID).Scan(&p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("project %s not found", p.ID)
		}
		return nil, fmt.Errorf("update project: %w", err)
	}
	return p, nil
}

// DeleteProject removes a project (cascade deletes members, persons, flags, etc.)
func (r *Repository) DeleteProject(ctx context.Context, id string) error {
	result, err := r.connManager.Primary().ExecContext(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("project %s not found", id)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Project Members
// ─────────────────────────────────────────────────────────────────────────────

// AddProjectMember adds or updates a membership record
func (r *Repository) AddProjectMember(ctx context.Context, m *dao.ProjectMember) (*dao.ProjectMember, error) {
	query := `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (project_id, user_id) DO UPDATE SET role = EXCLUDED.role
		RETURNING created_at
	`
	err := r.connManager.Primary().QueryRowContext(ctx, query, m.ProjectID, m.UserID, m.Role).Scan(&m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("add project member: %w", err)
	}
	return m, nil
}

// RemoveProjectMember removes a user from a project
func (r *Repository) RemoveProjectMember(ctx context.Context, projectID, userID string) error {
	result, err := r.connManager.Primary().ExecContext(ctx,
		`DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`, projectID, userID)
	if err != nil {
		return fmt.Errorf("remove project member: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("member %s not found in project %s", userID, projectID)
	}
	return nil
}

// GetProjectMembers returns all members of a project
func (r *Repository) GetProjectMembers(ctx context.Context, projectID string) ([]*dao.ProjectMember, error) {
	rows, err := r.connManager.Primary().QueryContext(ctx,
		`SELECT project_id, user_id, role, created_at FROM project_members WHERE project_id = $1`, projectID)
	if err != nil {
		return nil, fmt.Errorf("get project members: %w", err)
	}
	defer rows.Close()
	var members []*dao.ProjectMember
	for rows.Next() {
		var m dao.ProjectMember
		if err := rows.Scan(&m.ProjectID, &m.UserID, &m.Role, &m.CreatedAt); err != nil {
			logger.Error(ctx, "scan member row: %v", err)
			continue
		}
		members = append(members, &m)
	}
	return members, rows.Err()
}

// IsProjectMember checks if userID is a member of the given project
func (r *Repository) IsProjectMember(ctx context.Context, projectID, userID string) (bool, error) {
	var count int
	err := r.connManager.Primary().QueryRowContext(ctx,
		`SELECT COUNT(*) FROM project_members WHERE project_id = $1 AND user_id = $2`, projectID, userID,
	).Scan(&count)
	return count > 0, err
}

// ─────────────────────────────────────────────────────────────────────────────
// Persons
// ─────────────────────────────────────────────────────────────────────────────

// UpsertPerson creates or updates a person record (upsert on project_id+distinct_id)
func (r *Repository) UpsertPerson(ctx context.Context, p *dao.Person) (*dao.Person, error) {
	query := `
		INSERT INTO persons (project_id, distinct_id, properties)
		VALUES ($1, $2, $3)
		ON CONFLICT (project_id, distinct_id) DO UPDATE
		  SET properties = persons.properties || EXCLUDED.properties
		RETURNING id, created_at
	`
	propsJSON, err := json.Marshal(p.Properties)
	if err != nil {
		return nil, fmt.Errorf("marshal person properties: %w", err)
	}
	err = r.connManager.Primary().QueryRowContext(ctx, query,
		p.ProjectID, p.DistinctID, propsJSON,
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert person: %w", err)
	}
	return p, nil
}

// GetPerson fetches a person by distinct_id within a project
func (r *Repository) GetPerson(ctx context.Context, projectID, distinctID string) (*dao.Person, error) {
	var p dao.Person
	var propsJSON []byte
	err := r.connManager.Primary().QueryRowContext(ctx,
		`SELECT id, project_id, distinct_id, properties, created_at FROM persons WHERE project_id = $1 AND distinct_id = $2`,
		projectID, distinctID,
	).Scan(&p.ID, &p.ProjectID, &p.DistinctID, &propsJSON, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("person %s not found", distinctID)
		}
		return nil, fmt.Errorf("get person: %w", err)
	}
	p.Properties = propsJSON
	return &p, nil
}

// ListPersons returns paginated persons for a project with optional search
func (r *Repository) ListPersons(ctx context.Context, f *filter.PersonFilter) ([]*dao.Person, int, error) {
	base := `FROM persons WHERE project_id = $1`
	args := []interface{}{f.ProjectID}
	idx := 2

	if f.Search != "" {
		base += fmt.Sprintf(" AND (distinct_id ILIKE $%d OR properties->>'email' ILIKE $%d)", idx, idx)
		args = append(args, "%"+f.Search+"%")
		idx++
	}

	var total int
	if err := r.connManager.Primary().QueryRowContext(ctx, "SELECT COUNT(*) "+base, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count persons: %w", err)
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	args = append(args, limit, f.Offset)
	rows, err := r.connManager.Primary().QueryContext(ctx,
		fmt.Sprintf(`SELECT id, project_id, distinct_id, properties, created_at %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, base, idx, idx+1),
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list persons: %w", err)
	}
	defer rows.Close()

	var persons []*dao.Person
	for rows.Next() {
		var p dao.Person
		var propsJSON []byte
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.DistinctID, &propsJSON, &p.CreatedAt); err != nil {
			logger.Error(ctx, "scan person: %v", err)
			continue
		}
		p.Properties = propsJSON
		persons = append(persons, &p)
	}
	return persons, total, rows.Err()
}

// DeletePerson deletes a person by distinct_id within a project
func (r *Repository) DeletePerson(ctx context.Context, projectID, distinctID string) error {
	result, err := r.connManager.Primary().ExecContext(ctx,
		`DELETE FROM persons WHERE project_id = $1 AND distinct_id = $2`, projectID, distinctID)
	if err != nil {
		return fmt.Errorf("delete person: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("person %s not found", distinctID)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Feature Flags
// ─────────────────────────────────────────────────────────────────────────────

// CreateFeatureFlag inserts a new feature flag
func (r *Repository) CreateFeatureFlag(ctx context.Context, f *dao.FeatureFlag) (*dao.FeatureFlag, error) {
	filtersJSON, err := json.Marshal(f.Filters)
	if err != nil {
		return nil, fmt.Errorf("marshal filters: %w", err)
	}
	query := `
		INSERT INTO feature_flags (project_id, key, name, active, rollout_pct, filters)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	err = r.connManager.Primary().QueryRowContext(ctx, query,
		f.ProjectID, f.Key, f.Name, f.Active, f.RolloutPct, filtersJSON,
	).Scan(&f.ID, &f.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("feature flag with key '%s' already exists in project", f.Key)
		}
		return nil, fmt.Errorf("create feature flag: %w", err)
	}
	return f, nil
}

// GetFeatureFlagByID fetches a feature flag by ID
func (r *Repository) GetFeatureFlagByID(ctx context.Context, projectID, flagID string) (*dao.FeatureFlag, error) {
	var f dao.FeatureFlag
	var filtersJSON []byte
	err := r.connManager.Primary().QueryRowContext(ctx,
		`SELECT id, project_id, key, name, active, rollout_pct, filters, created_at FROM feature_flags WHERE id = $1 AND project_id = $2`,
		flagID, projectID,
	).Scan(&f.ID, &f.ProjectID, &f.Key, &f.Name, &f.Active, &f.RolloutPct, &filtersJSON, &f.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("feature flag %s not found", flagID)
		}
		return nil, fmt.Errorf("get feature flag: %w", err)
	}
	f.Filters = filtersJSON
	return &f, nil
}

// ListFeatureFlags returns all flags for a project
func (r *Repository) ListFeatureFlags(ctx context.Context, f *filter.FeatureFlagFilter) ([]*dao.FeatureFlag, error) {
	rows, err := r.connManager.Primary().QueryContext(ctx,
		`SELECT id, project_id, key, name, active, rollout_pct, filters, created_at FROM feature_flags WHERE project_id = $1 ORDER BY created_at DESC`,
		f.ProjectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list feature flags: %w", err)
	}
	defer rows.Close()
	var flags []*dao.FeatureFlag
	for rows.Next() {
		var flag dao.FeatureFlag
		var filtersJSON []byte
		if err := rows.Scan(&flag.ID, &flag.ProjectID, &flag.Key, &flag.Name, &flag.Active, &flag.RolloutPct, &filtersJSON, &flag.CreatedAt); err != nil {
			logger.Error(ctx, "scan feature flag: %v", err)
			continue
		}
		flag.Filters = filtersJSON
		flags = append(flags, &flag)
	}
	return flags, rows.Err()
}

// ListActiveFeatureFlags returns all active flags for a project (used for /decide/ evaluation)
func (r *Repository) ListActiveFeatureFlags(ctx context.Context, projectID string) ([]*dao.FeatureFlag, error) {
	rows, err := r.connManager.Primary().QueryContext(ctx,
		`SELECT id, project_id, key, name, active, rollout_pct, filters, created_at FROM feature_flags WHERE project_id = $1 AND active = true`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list active feature flags: %w", err)
	}
	defer rows.Close()
	var flags []*dao.FeatureFlag
	for rows.Next() {
		var flag dao.FeatureFlag
		var filtersJSON []byte
		if err := rows.Scan(&flag.ID, &flag.ProjectID, &flag.Key, &flag.Name, &flag.Active, &flag.RolloutPct, &filtersJSON, &flag.CreatedAt); err != nil {
			logger.Error(ctx, "scan active flag: %v", err)
			continue
		}
		flag.Filters = filtersJSON
		flags = append(flags, &flag)
	}
	return flags, rows.Err()
}

// UpdateFeatureFlag applies partial updates to a feature flag
func (r *Repository) UpdateFeatureFlag(ctx context.Context, f *dao.FeatureFlag) (*dao.FeatureFlag, error) {
	filtersJSON, err := json.Marshal(f.Filters)
	if err != nil {
		return nil, fmt.Errorf("marshal filters: %w", err)
	}
	query := `
		UPDATE feature_flags SET name = $1, active = $2, rollout_pct = $3, filters = $4
		WHERE id = $5 AND project_id = $6
		RETURNING created_at
	`
	err = r.connManager.Primary().QueryRowContext(ctx, query,
		f.Name, f.Active, f.RolloutPct, filtersJSON, f.ID, f.ProjectID,
	).Scan(&f.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("feature flag %s not found", f.ID)
		}
		return nil, fmt.Errorf("update feature flag: %w", err)
	}
	return f, nil
}

// DeleteFeatureFlag removes a feature flag
func (r *Repository) DeleteFeatureFlag(ctx context.Context, projectID, flagID string) error {
	result, err := r.connManager.Primary().ExecContext(ctx,
		`DELETE FROM feature_flags WHERE id = $1 AND project_id = $2`, flagID, projectID)
	if err != nil {
		return fmt.Errorf("delete feature flag: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("feature flag %s not found", flagID)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Dashboards
// ─────────────────────────────────────────────────────────────────────────────

// CreateDashboard inserts a new dashboard
func (r *Repository) CreateDashboard(ctx context.Context, d *dao.Dashboard) (*dao.Dashboard, error) {
	layoutJSON, _ := json.Marshal(d.Layout)
	err := r.connManager.Primary().QueryRowContext(ctx,
		`INSERT INTO dashboards (project_id, name, layout) VALUES ($1, $2, $3) RETURNING id, created_at`,
		d.ProjectID, d.Name, layoutJSON,
	).Scan(&d.ID, &d.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create dashboard: %w", err)
	}
	return d, nil
}

// GetDashboard fetches a dashboard with all its items
func (r *Repository) GetDashboard(ctx context.Context, projectID, dashboardID string) (*dao.Dashboard, []*dao.DashboardItem, error) {
	var d dao.Dashboard
	var layoutJSON []byte
	err := r.connManager.Primary().QueryRowContext(ctx,
		`SELECT id, project_id, name, layout, created_at FROM dashboards WHERE id = $1 AND project_id = $2`,
		dashboardID, projectID,
	).Scan(&d.ID, &d.ProjectID, &d.Name, &layoutJSON, &d.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, fmt.Errorf("dashboard %s not found", dashboardID)
		}
		return nil, nil, fmt.Errorf("get dashboard: %w", err)
	}
	d.Layout = layoutJSON

	rows, err := r.connManager.Primary().QueryContext(ctx,
		`SELECT id, dashboard_id, name, type, query, position FROM dashboard_items WHERE dashboard_id = $1 ORDER BY id`,
		dashboardID,
	)
	if err != nil {
		return &d, nil, fmt.Errorf("get dashboard items: %w", err)
	}
	defer rows.Close()
	var items []*dao.DashboardItem
	for rows.Next() {
		var item dao.DashboardItem
		var queryJSON, posJSON []byte
		if err := rows.Scan(&item.ID, &item.DashboardID, &item.Name, &item.Type, &queryJSON, &posJSON); err != nil {
			logger.Error(ctx, "scan dashboard item: %v", err)
			continue
		}
		item.Query = queryJSON
		item.Position = posJSON
		items = append(items, &item)
	}
	return &d, items, rows.Err()
}

// ListDashboards returns all dashboards for a project with item count
func (r *Repository) ListDashboards(ctx context.Context, f *filter.DashboardFilter) ([]*dao.Dashboard, []int, error) {
	rows, err := r.connManager.Primary().QueryContext(ctx, `
		SELECT d.id, d.project_id, d.name, d.layout, d.created_at,
		       COUNT(di.id) AS item_count
		FROM dashboards d
		LEFT JOIN dashboard_items di ON d.id = di.dashboard_id
		WHERE d.project_id = $1
		GROUP BY d.id
		ORDER BY d.created_at DESC
	`, f.ProjectID)
	if err != nil {
		return nil, nil, fmt.Errorf("list dashboards: %w", err)
	}
	defer rows.Close()
	var dashboards []*dao.Dashboard
	var counts []int
	for rows.Next() {
		var d dao.Dashboard
		var layoutJSON []byte
		var cnt int
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.Name, &layoutJSON, &d.CreatedAt, &cnt); err != nil {
			logger.Error(ctx, "scan dashboard: %v", err)
			continue
		}
		d.Layout = layoutJSON
		dashboards = append(dashboards, &d)
		counts = append(counts, cnt)
	}
	return dashboards, counts, rows.Err()
}

// DeleteDashboard removes a dashboard and all its items
func (r *Repository) DeleteDashboard(ctx context.Context, projectID, dashboardID string) error {
	result, err := r.connManager.Primary().ExecContext(ctx,
		`DELETE FROM dashboards WHERE id = $1 AND project_id = $2`, dashboardID, projectID)
	if err != nil {
		return fmt.Errorf("delete dashboard: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("dashboard %s not found", dashboardID)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Dashboard Items
// ─────────────────────────────────────────────────────────────────────────────

// CreateDashboardItem adds an insight panel to a dashboard
func (r *Repository) CreateDashboardItem(ctx context.Context, item *dao.DashboardItem) (*dao.DashboardItem, error) {
	queryJSON, _ := json.Marshal(item.Query)
	posJSON, _ := json.Marshal(item.Position)
	err := r.connManager.Primary().QueryRowContext(ctx,
		`INSERT INTO dashboard_items (dashboard_id, name, type, query, position) VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		item.DashboardID, item.Name, item.Type, queryJSON, posJSON,
	).Scan(&item.ID)
	if err != nil {
		return nil, fmt.Errorf("create dashboard item: %w", err)
	}
	return item, nil
}

// UpdateDashboardItem updates a dashboard item's name, query, or position
func (r *Repository) UpdateDashboardItem(ctx context.Context, item *dao.DashboardItem) (*dao.DashboardItem, error) {
	queryJSON, _ := json.Marshal(item.Query)
	posJSON, _ := json.Marshal(item.Position)
	result, err := r.connManager.Primary().ExecContext(ctx,
		`UPDATE dashboard_items SET name=$1, type=$2, query=$3, position=$4 WHERE id=$5 AND dashboard_id=$6`,
		item.Name, item.Type, queryJSON, posJSON, item.ID, item.DashboardID,
	)
	if err != nil {
		return nil, fmt.Errorf("update dashboard item: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return nil, fmt.Errorf("dashboard item %s not found", item.ID)
	}
	return item, nil
}

// DeleteDashboardItem removes a single dashboard item
func (r *Repository) DeleteDashboardItem(ctx context.Context, dashboardID, itemID string) error {
	result, err := r.connManager.Primary().ExecContext(ctx,
		`DELETE FROM dashboard_items WHERE id = $1 AND dashboard_id = $2`, itemID, dashboardID)
	if err != nil {
		return fmt.Errorf("delete dashboard item: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("dashboard item %s not found", itemID)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Cohorts
// ─────────────────────────────────────────────────────────────────────────────

// cohortFilterCfg mirrors the JSON structure stored in cohorts.filters.
type cohortFilterCfg struct {
	Groups []struct {
		Properties []struct {
			Key      string `json:"key"`
			Value    any    `json:"value"`
			Operator string `json:"operator"`
			Type     string `json:"type"`
		} `json:"properties"`
	} `json:"groups"`
}

// countPersonsMatchingFilters counts persons in a project that satisfy the
// cohort's filter conditions. Groups are OR-combined; properties within a
// group are AND-combined.
func (r *Repository) countPersonsMatchingFilters(ctx context.Context, projectID string, filtersJSON []byte) (int, error) {
	baseCount := func() (int, error) {
		var n int
		err := r.connManager.Primary().QueryRowContext(ctx,
			`SELECT COUNT(*) FROM persons WHERE project_id = $1`, projectID).Scan(&n)
		return n, err
	}

	if len(filtersJSON) == 0 {
		return baseCount()
	}
	s := strings.TrimSpace(string(filtersJSON))
	if s == "{}" || s == "null" || s == "" {
		return baseCount()
	}

	var cfg cohortFilterCfg
	if err := json.Unmarshal(filtersJSON, &cfg); err != nil || len(cfg.Groups) == 0 {
		return baseCount()
	}

	args := []interface{}{projectID}
	idx := 2
	var groupClauses []string

	for _, group := range cfg.Groups {
		var propClauses []string
		for _, prop := range group.Properties {
			// Reject keys that could break out of the jsonb operator expression.
			if !isSafePropertyKey(prop.Key) {
				continue
			}
			switch prop.Operator {
			case "exact":
				propClauses = append(propClauses, fmt.Sprintf("properties->>'%s' = $%d", prop.Key, idx))
				args = append(args, fmt.Sprintf("%v", prop.Value))
				idx++
			case "contains":
				propClauses = append(propClauses, fmt.Sprintf("properties->>'%s' ILIKE $%d", prop.Key, idx))
				args = append(args, "%"+fmt.Sprintf("%v", prop.Value)+"%")
				idx++
			case "gt":
				propClauses = append(propClauses, fmt.Sprintf("(properties->>'%s')::numeric > $%d", prop.Key, idx))
				args = append(args, prop.Value)
				idx++
			case "lt":
				propClauses = append(propClauses, fmt.Sprintf("(properties->>'%s')::numeric < $%d", prop.Key, idx))
				args = append(args, prop.Value)
				idx++
			}
		}
		if len(propClauses) > 0 {
			groupClauses = append(groupClauses, "("+strings.Join(propClauses, " AND ")+")")
		}
	}

	where := "project_id = $1"
	if len(groupClauses) > 0 {
		where += " AND (" + strings.Join(groupClauses, " OR ") + ")"
	}

	var count int
	err := r.connManager.Primary().QueryRowContext(ctx,
		"SELECT COUNT(*) FROM persons WHERE "+where, args...).Scan(&count)
	return count, err
}

// isSafePropertyKey ensures the key contains only characters safe to embed in
// a JSONB operator expression (properties->>'key').
func isSafePropertyKey(key string) bool {
	if key == "" {
		return false
	}
	for _, c := range key {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.' || c == '@' || c == '$') {
			return false
		}
	}
	return true
}

// CreateCohort inserts a new cohort
func (r *Repository) CreateCohort(ctx context.Context, c *dao.Cohort) (*dao.Cohort, error) {
	filtersJSON, _ := json.Marshal(c.Filters)
	err := r.connManager.Primary().QueryRowContext(ctx,
		`INSERT INTO cohorts (project_id, name, filters) VALUES ($1,$2,$3) RETURNING id, created_at`,
		c.ProjectID, c.Name, filtersJSON,
	).Scan(&c.ID, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create cohort: %w", err)
	}
	count, err := r.countPersonsMatchingFilters(ctx, c.ProjectID, filtersJSON)
	if err != nil {
		logger.Error(ctx, "count persons for cohort: %v", err)
	}
	c.PersonCount = count
	return c, nil
}

// ListCohorts returns all cohorts for a project
func (r *Repository) ListCohorts(ctx context.Context, f *filter.CohortFilter) ([]*dao.Cohort, error) {
	rows, err := r.connManager.Primary().QueryContext(ctx,
		`SELECT id, project_id, name, filters, created_at FROM cohorts WHERE project_id = $1 ORDER BY created_at DESC`,
		f.ProjectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list cohorts: %w", err)
	}
	defer rows.Close()
	var cohorts []*dao.Cohort
	for rows.Next() {
		var c dao.Cohort
		var filtersJSON []byte
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.Name, &filtersJSON, &c.CreatedAt); err != nil {
			logger.Error(ctx, "scan cohort: %v", err)
			continue
		}
		c.Filters = filtersJSON
		count, err := r.countPersonsMatchingFilters(ctx, c.ProjectID, filtersJSON)
		if err != nil {
			logger.Error(ctx, "count persons for cohort %s: %v", c.ID, err)
		}
		c.PersonCount = count
		cohorts = append(cohorts, &c)
	}
	return cohorts, rows.Err()
}

// DeleteCohort removes a cohort
func (r *Repository) DeleteCohort(ctx context.Context, projectID, cohortID string) error {
	result, err := r.connManager.Primary().ExecContext(ctx,
		`DELETE FROM cohorts WHERE id = $1 AND project_id = $2`, cohortID, projectID)
	if err != nil {
		return fmt.Errorf("delete cohort: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("cohort %s not found", cohortID)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// isUniqueViolation checks if the error is a PostgreSQL unique constraint violation
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "unique constraint") || strings.Contains(msg, "duplicate key")
}

// ensure time package is used
var _ = time.Now
