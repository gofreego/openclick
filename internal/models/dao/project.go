package dao

import "time"

// Project represents a tracked application (equivalent to PostHog team)
type Project struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	APIKey    string    `db:"api_key"`    // public key for SDK ingestion: ock_pub_...
	SecretKey string    `db:"secret_key"` // private key for server-side APIs: ock_sec_...
	Timezone  string    `db:"timezone"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// ProjectMember represents membership of an external user in a project
type ProjectMember struct {
	ProjectID string    `db:"project_id"`
	UserID    string    `db:"user_id"` // external user ID from auth service
	Role      string    `db:"role"`    // "owner" | "member"
	CreatedAt time.Time `db:"created_at"`
}
