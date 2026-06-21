-- Drop PostgreSQL tables in reverse dependency order

DROP TABLE IF EXISTS annotations;
DROP TABLE IF EXISTS dashboard_items;
DROP TABLE IF EXISTS dashboards;
DROP TABLE IF EXISTS feature_flags;
DROP TABLE IF EXISTS cohorts;
DROP TABLE IF EXISTS persons;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;
