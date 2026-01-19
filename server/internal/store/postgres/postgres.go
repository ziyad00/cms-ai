package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/ziyad/cms-ai/server/internal/store"
)

type PostgresStore struct {
	db *sql.DB
}

func New(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	store := &PostgresStore{db: db}

	// Auto-run migrations on startup
	if err := store.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

func (p *PostgresStore) Close() error {
	return p.db.Close()
}

// runMigrations executes SQL migration scripts to set up the database schema
func (p *PostgresStore) runMigrations() error {
	// Check if users table exists to determine if migrations are needed
	var exists bool
	err := p.db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if migrations are needed: %w", err)
	}

	if exists {
		// Tables already exist, skip migrations
		return nil
	}

	// Run initial migration SQL
	migrationSQL := `
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

		CREATE TABLE organizations (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE TABLE users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			email TEXT UNIQUE NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE TABLE memberships (
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
			role TEXT NOT NULL CHECK (role IN ('Owner', 'Admin', 'Editor', 'Viewer')),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (user_id, org_id)
		);

		CREATE TABLE brand_kits (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			tokens JSONB,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE TABLE templates (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
			owner_user_id UUID REFERENCES users(id),
			name TEXT NOT NULL,
			status TEXT NOT NULL CHECK (status IN ('Draft', 'Published', 'Archived')),
			current_version_id UUID,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			latest_version_no INT DEFAULT 0
		);

		CREATE TABLE template_versions (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			template_id UUID REFERENCES templates(id) ON DELETE CASCADE,
			org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
			version_no INT NOT NULL,
			spec_json JSONB NOT NULL,
			created_by UUID REFERENCES users(id),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE (template_id, version_no)
		);

		ALTER TABLE templates ADD CONSTRAINT fk_current_version FOREIGN KEY (current_version_id) REFERENCES template_versions(id);

		CREATE TABLE assets (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
			type TEXT NOT NULL CHECK (type IN ('pptx', 'png', 'file')),
			path TEXT NOT NULL,
			mime TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE TABLE jobs (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
			type TEXT NOT NULL CHECK (type IN ('render', 'preview', 'export')),
			status TEXT NOT NULL CHECK (status IN ('Queued', 'Running', 'Done', 'Failed', 'Retry', 'DeadLetter')),
			input_ref TEXT NOT NULL,
			output_ref TEXT,
			error TEXT,
			retry_count INT DEFAULT 0,
			max_retries INT DEFAULT 3,
			last_retry_at TIMESTAMPTZ,
			deduplication_id TEXT,
			metadata JSONB,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE TABLE metering_events (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(id),
			event_type TEXT NOT NULL,
			quantity INT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE TABLE audit_logs (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
			actor_user_id UUID REFERENCES users(id),
			action TEXT NOT NULL,
			target_ref TEXT,
			metadata JSONB,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		-- Auth update migration (002)
		ALTER TABLE users ADD COLUMN name TEXT;
		ALTER TABLE users ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
		ALTER TABLE memberships RENAME TO user_orgs;

		-- Indexes for common queries
		CREATE INDEX idx_templates_org ON templates(org_id);
		CREATE INDEX idx_template_versions_org ON template_versions(org_id);
		CREATE INDEX idx_assets_org ON assets(org_id);
		CREATE INDEX idx_jobs_org ON jobs(org_id);
		CREATE INDEX idx_metering_org_type ON metering_events(org_id, event_type);
		CREATE INDEX idx_audit_org_action ON audit_logs(org_id, action);
	`

	_, err = p.db.Exec(migrationSQL)
	if err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	return nil
}

func (p *PostgresStore) Templates() store.TemplateStore { return (*postgresTemplateStore)(p) }
func (p *PostgresStore) BrandKits() store.BrandKitStore { return (*postgresBrandKitStore)(p) }
func (p *PostgresStore) Assets() store.AssetStore       { return (*postgresAssetStore)(p) }
func (p *PostgresStore) Jobs() store.JobStore           { return (*postgresJobStore)(p) }
func (p *PostgresStore) Metering() store.MeteringStore  { return (*postgresMeteringStore)(p) }
func (p *PostgresStore) Audit() store.AuditStore        { return (*postgresAuditStore)(p) }
func (p *PostgresStore) Users() store.UserStore         { return (*postgresUserStore)(p) }
func (p *PostgresStore) Organizations() store.OrganizationStore {
	return (*postgresOrganizationStore)(p)
}

type postgresTemplateStore PostgresStore

// Implement basic CreateTemplate and ListTemplates for demo
func (p *postgresTemplateStore) CreateTemplate(ctx context.Context, t store.Template) (store.Template, error) {
	ps := (*PostgresStore)(p)
	query := `
		INSERT INTO templates (id, org_id, owner_user_id, name, status, latest_version_no, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	if t.ID == "" {
		t.ID = fmt.Sprintf("tpl-%s", generateID())
	}
	_, err := ps.db.ExecContext(ctx, query, t.ID, t.OrgID, t.OwnerUserID, t.Name, t.Status, t.LatestVersionNo, t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return store.Template{}, err
	}
	return t, nil
}

func (p *postgresTemplateStore) ListTemplates(ctx context.Context, orgID string) ([]store.Template, error) {
	ps := (*PostgresStore)(p)
	query := `SELECT id, org_id, owner_user_id, name, status, current_version_id, created_at, updated_at, latest_version_no FROM templates WHERE org_id = $1`
	rows, err := ps.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ts []store.Template
	for rows.Next() {
		var t store.Template
		err := rows.Scan(&t.ID, &t.OrgID, &t.OwnerUserID, &t.Name, &t.Status, &t.CurrentVersion, &t.CreatedAt, &t.UpdatedAt, &t.LatestVersionNo)
		if err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	return ts, nil
}

func (p *postgresTemplateStore) GetTemplate(ctx context.Context, orgID, id string) (store.Template, bool, error) {
	ps := (*PostgresStore)(p)
	query := `SELECT id, org_id, owner_user_id, name, status, current_version_id, created_at, updated_at, latest_version_no FROM templates WHERE org_id = $1 AND id = $2`
	var t store.Template
	err := ps.db.QueryRowContext(ctx, query, orgID, id).Scan(&t.ID, &t.OrgID, &t.OwnerUserID, &t.Name, &t.Status, &t.CurrentVersion, &t.CreatedAt, &t.UpdatedAt, &t.LatestVersionNo)
	if err != nil {
		if err == sql.ErrNoRows {
			return store.Template{}, false, nil
		}
		return store.Template{}, false, err
	}
	return t, true, nil
}

func (p *postgresTemplateStore) UpdateTemplate(ctx context.Context, t store.Template) (store.Template, error) {
	ps := (*PostgresStore)(p)
	query := `UPDATE templates SET name = $1, status = $2, current_version_id = $3, updated_at = $4, latest_version_no = $5 WHERE id = $6 AND org_id = $7`
	t.UpdatedAt = time.Now().UTC()
	_, err := ps.db.ExecContext(ctx, query, t.Name, t.Status, t.CurrentVersion, t.UpdatedAt, t.LatestVersionNo, t.ID, t.OrgID)
	if err != nil {
		return store.Template{}, err
	}
	return t, nil
}

func (p *postgresTemplateStore) CreateVersion(ctx context.Context, v store.TemplateVersion) (store.TemplateVersion, error) {
	ps := (*PostgresStore)(p)
	query := `INSERT INTO template_versions (id, template_id, org_id, version_no, spec_json, created_by, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	if v.ID == "" {
		v.ID = fmt.Sprintf("ver-%s", generateID())
	}
	v.CreatedAt = time.Now().UTC()
	_, err := ps.db.ExecContext(ctx, query, v.ID, v.Template, v.OrgID, v.VersionNo, v.SpecJSON, v.CreatedBy, v.CreatedAt)
	if err != nil {
		return store.TemplateVersion{}, err
	}
	return v, nil
}

func (p *postgresTemplateStore) ListVersions(ctx context.Context, orgID, templateID string) ([]store.TemplateVersion, error) {
	ps := (*PostgresStore)(p)
	query := `SELECT id, template_id, org_id, version_no, spec_json, created_by, created_at FROM template_versions WHERE org_id = $1 AND template_id = $2 ORDER BY version_no DESC`
	rows, err := ps.db.QueryContext(ctx, query, orgID, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vs []store.TemplateVersion
	for rows.Next() {
		var v store.TemplateVersion
		err := rows.Scan(&v.ID, &v.Template, &v.OrgID, &v.VersionNo, &v.SpecJSON, &v.CreatedBy, &v.CreatedAt)
		if err != nil {
			return nil, err
		}
		vs = append(vs, v)
	}
	return vs, nil
}

func (p *postgresTemplateStore) GetVersion(ctx context.Context, orgID, versionID string) (store.TemplateVersion, bool, error) {
	ps := (*PostgresStore)(p)
	query := `SELECT id, template_id, org_id, version_no, spec_json, created_by, created_at FROM template_versions WHERE org_id = $1 AND id = $2`
	var v store.TemplateVersion
	err := ps.db.QueryRowContext(ctx, query, orgID, versionID).Scan(&v.ID, &v.Template, &v.OrgID, &v.VersionNo, &v.SpecJSON, &v.CreatedBy, &v.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return store.TemplateVersion{}, false, nil
		}
		return store.TemplateVersion{}, false, err
	}
	return v, true, nil
}

// Other stores are stubs
type postgresBrandKitStore PostgresStore

func (p *postgresBrandKitStore) Create(ctx context.Context, b store.BrandKit) (store.BrandKit, error) {
	ps := (*PostgresStore)(p)
	query := `INSERT INTO brand_kits (id, org_id, name, tokens, created_at) VALUES ($1, $2, $3, $4, $5)`
	if b.ID == "" {
		b.ID = fmt.Sprintf("bk-%s", generateID())
	}
	b.CreatedAt = time.Now().UTC()
	_, err := ps.db.ExecContext(ctx, query, b.ID, b.OrgID, b.Name, b.Tokens, b.CreatedAt)
	return b, err
}

func (p *postgresBrandKitStore) List(ctx context.Context, orgID string) ([]store.BrandKit, error) {
	ps := (*PostgresStore)(p)
	query := `SELECT id, org_id, name, tokens, created_at FROM brand_kits WHERE org_id = $1`
	rows, err := ps.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bks []store.BrandKit
	for rows.Next() {
		var b store.BrandKit
		err := rows.Scan(&b.ID, &b.OrgID, &b.Name, &b.Tokens, &b.CreatedAt)
		if err != nil {
			return nil, err
		}
		bks = append(bks, b)
	}
	return bks, nil
}

type postgresAssetStore PostgresStore

func (p *postgresAssetStore) Create(ctx context.Context, a store.Asset) (store.Asset, error) {
	ps := (*PostgresStore)(p)
	query := `INSERT INTO assets (id, org_id, type, path, mime, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	if a.ID == "" {
		a.ID = fmt.Sprintf("asset-%s", generateID())
	}
	a.CreatedAt = time.Now().UTC()
	_, err := ps.db.ExecContext(ctx, query, a.ID, a.OrgID, a.Type, a.Path, a.Mime, a.CreatedAt)
	return a, err
}

func (p *postgresAssetStore) Get(ctx context.Context, orgID, id string) (store.Asset, bool, error) {
	ps := (*PostgresStore)(p)
	query := `SELECT id, org_id, type, path, mime, created_at FROM assets WHERE org_id = $1 AND id = $2`
	var a store.Asset
	err := ps.db.QueryRowContext(ctx, query, orgID, id).Scan(&a.ID, &a.OrgID, &a.Type, &a.Path, &a.Mime, &a.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return store.Asset{}, false, nil
		}
		return store.Asset{}, false, err
	}
	return a, true, nil
}

func (p *postgresAssetStore) Store(_ context.Context, orgID, assetID string, data []byte) (string, error) {
	// For now, store to local filesystem in a data directory
	// In production, this would use object storage like S3
	dataDir := "data/assets"
	orgDir := filepath.Join(dataDir, orgID)

	if err := os.MkdirAll(orgDir, 0o755); err != nil {
		return "", err
	}

	filePath := filepath.Join(orgDir, assetID)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return "", err
	}

	return filePath, nil
}

type postgresJobStore PostgresStore

func (p *postgresJobStore) Enqueue(ctx context.Context, j store.Job) (store.Job, error) {
	ps := (*PostgresStore)(p)
	query := `INSERT INTO jobs (id, org_id, type, status, input_ref, output_ref, error, retry_count, max_retries, last_retry_at, deduplication_id, metadata, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	if j.ID == "" {
		j.ID = fmt.Sprintf("job-%s", generateID())
	}
	if j.MaxRetries == 0 {
		j.MaxRetries = 3
	}
	j.CreatedAt = time.Now().UTC()
	j.UpdatedAt = j.CreatedAt
	_, err := ps.db.ExecContext(ctx, query, j.ID, j.OrgID, j.Type, j.Status, j.InputRef, j.OutputRef, j.Error, j.RetryCount, j.MaxRetries, j.LastRetryAt, j.DeduplicationID, j.Metadata, j.CreatedAt, j.UpdatedAt)
	return j, err
}

func (p *postgresJobStore) EnqueueWithDeduplication(ctx context.Context, j store.Job) (store.Job, bool, error) {
	ps := (*PostgresStore)(p)

	if j.DeduplicationID != "" {
		query := `SELECT id, org_id, type, status, input_ref, output_ref, error, retry_count, max_retries, last_retry_at, deduplication_id, metadata, created_at, updated_at FROM jobs WHERE org_id = $1 AND deduplication_id = $2 AND status IN ('Queued', 'Running')`
		var existingJob store.Job
		err := ps.db.QueryRowContext(ctx, query, j.OrgID, j.DeduplicationID).Scan(
			&existingJob.ID, &existingJob.OrgID, &existingJob.Type, &existingJob.Status,
			&existingJob.InputRef, &existingJob.OutputRef, &existingJob.Error,
			&existingJob.RetryCount, &existingJob.MaxRetries, &existingJob.LastRetryAt,
			&existingJob.DeduplicationID, &existingJob.Metadata, &existingJob.CreatedAt, &existingJob.UpdatedAt,
		)
		if err == nil {
			return existingJob, true, nil
		}
		if err != sql.ErrNoRows {
			return store.Job{}, false, err
		}
	}

	inserted, err := p.Enqueue(ctx, j)
	return inserted, false, err
}

func (p *postgresJobStore) Get(ctx context.Context, orgID, jobID string) (store.Job, bool, error) {
	ps := (*PostgresStore)(p)
	query := `SELECT id, org_id, type, status, input_ref, output_ref, error, retry_count, max_retries, last_retry_at, deduplication_id, metadata, created_at, updated_at FROM jobs WHERE org_id = $1 AND id = $2`
	var j store.Job
	err := ps.db.QueryRowContext(ctx, query, orgID, jobID).Scan(&j.ID, &j.OrgID, &j.Type, &j.Status, &j.InputRef, &j.OutputRef, &j.Error, &j.RetryCount, &j.MaxRetries, &j.LastRetryAt, &j.DeduplicationID, &j.Metadata, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return store.Job{}, false, nil
		}
		return store.Job{}, false, err
	}
	return j, true, nil
}

func (p *postgresJobStore) GetByDeduplicationID(ctx context.Context, orgID, dedupID string) (store.Job, bool, error) {
	ps := (*PostgresStore)(p)
	query := `SELECT id, org_id, type, status, input_ref, output_ref, error, retry_count, max_retries, last_retry_at, deduplication_id, metadata, created_at, updated_at FROM jobs WHERE org_id = $1 AND deduplication_id = $2`
	var j store.Job
	err := ps.db.QueryRowContext(ctx, query, orgID, dedupID).Scan(&j.ID, &j.OrgID, &j.Type, &j.Status, &j.InputRef, &j.OutputRef, &j.Error, &j.RetryCount, &j.MaxRetries, &j.LastRetryAt, &j.DeduplicationID, &j.Metadata, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return store.Job{}, false, nil
		}
		return store.Job{}, false, err
	}
	return j, true, nil
}

func (p *postgresJobStore) Update(ctx context.Context, j store.Job) (store.Job, error) {
	ps := (*PostgresStore)(p)
	query := `UPDATE jobs SET status = $1, output_ref = $2, error = $3, retry_count = $4, max_retries = $5, last_retry_at = $6, deduplication_id = $7, metadata = $8, updated_at = $9 WHERE id = $10 AND org_id = $11`
	j.UpdatedAt = time.Now().UTC()
	_, err := ps.db.ExecContext(ctx, query, j.Status, j.OutputRef, j.Error, j.RetryCount, j.MaxRetries, j.LastRetryAt, j.DeduplicationID, j.Metadata, j.UpdatedAt, j.ID, j.OrgID)
	return j, err
}

func (p *postgresJobStore) ListQueued(ctx context.Context) ([]store.Job, error) {
	ps := (*PostgresStore)(p)
	query := `SELECT id, org_id, type, status, input_ref, output_ref, error, retry_count, max_retries, last_retry_at, deduplication_id, metadata, created_at, updated_at FROM jobs WHERE status = $1 ORDER BY created_at ASC`
	rows, err := ps.db.QueryContext(ctx, query, store.JobQueued)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []store.Job
	for rows.Next() {
		var job store.Job
		err := rows.Scan(&job.ID, &job.OrgID, &job.Type, &job.Status, &job.InputRef, &job.OutputRef, &job.Error, &job.RetryCount, &job.MaxRetries, &job.LastRetryAt, &job.DeduplicationID, &job.Metadata, &job.CreatedAt, &job.UpdatedAt)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func (p *postgresJobStore) ListRetry(ctx context.Context) ([]store.Job, error) {
	ps := (*PostgresStore)(p)
	query := `SELECT id, org_id, type, status, input_ref, output_ref, error, retry_count, max_retries, last_retry_at, deduplication_id, metadata, created_at, updated_at FROM jobs WHERE status = $1 ORDER BY last_retry_at ASC`
	rows, err := ps.db.QueryContext(ctx, query, store.JobRetry)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []store.Job
	for rows.Next() {
		var job store.Job
		err := rows.Scan(&job.ID, &job.OrgID, &job.Type, &job.Status, &job.InputRef, &job.OutputRef, &job.Error, &job.RetryCount, &job.MaxRetries, &job.LastRetryAt, &job.DeduplicationID, &job.Metadata, &job.CreatedAt, &job.UpdatedAt)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func (p *postgresJobStore) ListDeadLetter(ctx context.Context) ([]store.Job, error) {
	ps := (*PostgresStore)(p)
	query := `SELECT id, org_id, type, status, input_ref, output_ref, error, retry_count, max_retries, last_retry_at, deduplication_id, metadata, created_at, updated_at FROM jobs WHERE status = $1 ORDER BY updated_at DESC`
	rows, err := ps.db.QueryContext(ctx, query, store.JobDeadLetter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []store.Job
	for rows.Next() {
		var job store.Job
		err := rows.Scan(&job.ID, &job.OrgID, &job.Type, &job.Status, &job.InputRef, &job.OutputRef, &job.Error, &job.RetryCount, &job.MaxRetries, &job.LastRetryAt, &job.DeduplicationID, &job.Metadata, &job.CreatedAt, &job.UpdatedAt)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func (p *postgresJobStore) MoveToDeadLetter(ctx context.Context, jobID string) error {
	ps := (*PostgresStore)(p)
	query := `UPDATE jobs SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := ps.db.ExecContext(ctx, query, store.JobDeadLetter, time.Now().UTC(), jobID)
	return err
}

func (p *postgresJobStore) RetryDeadLetterJob(ctx context.Context, jobID string) error {
	ps := (*PostgresStore)(p)
	query := `UPDATE jobs SET status = $1, retry_count = 0, error = NULL, updated_at = $2 WHERE id = $3`
	_, err := ps.db.ExecContext(ctx, query, store.JobQueued, time.Now().UTC(), jobID)
	return err
}

type postgresMeteringStore PostgresStore

func (p *postgresMeteringStore) Record(ctx context.Context, e store.MeteringEvent) (store.MeteringEvent, error) {
	ps := (*PostgresStore)(p)
	query := `INSERT INTO metering_events (id, org_id, user_id, event_type, quantity, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	if e.ID == "" {
		e.ID = fmt.Sprintf("met-%s", generateID())
	}
	e.CreatedAt = time.Now().UTC()
	_, err := ps.db.ExecContext(ctx, query, e.ID, e.OrgID, e.UserID, e.Type, e.Quantity, e.CreatedAt)
	return e, err
}

func (p *postgresMeteringStore) SumByType(ctx context.Context, orgID string, eventType string) (int, error) {
	ps := (*PostgresStore)(p)
	query := `SELECT COALESCE(SUM(quantity), 0) FROM metering_events WHERE org_id = $1 AND event_type = $2`
	var sum int
	err := ps.db.QueryRowContext(ctx, query, orgID, eventType).Scan(&sum)
	return sum, err
}

type postgresAuditStore PostgresStore

func (p *postgresAuditStore) Append(ctx context.Context, a store.AuditLog) (store.AuditLog, error) {
	ps := (*PostgresStore)(p)
	query := `INSERT INTO audit_logs (id, org_id, actor_user_id, action, target_ref, metadata, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	if a.ID == "" {
		a.ID = fmt.Sprintf("aud-%s", generateID())
	}
	a.CreatedAt = time.Now().UTC()
	_, err := ps.db.ExecContext(ctx, query, a.ID, a.OrgID, a.ActorID, a.Action, a.TargetRef, a.Metadata, a.CreatedAt)
	return a, err
}

type postgresUserStore PostgresStore

func (p *postgresUserStore) CreateUser(ctx context.Context, u *store.User) error {
	ps := (*PostgresStore)(p)
	// Let PostgreSQL generate the UUID automatically
	query := `INSERT INTO users (email, name) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	err := ps.db.QueryRowContext(ctx, query, u.Email, u.Name).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	return err
}

func (p *postgresUserStore) GetUser(ctx context.Context, userID string) (store.User, bool, error) {
	ps := (*PostgresStore)(p)
	query := `SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`
	var u store.User
	err := ps.db.QueryRowContext(ctx, query, userID).Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return u, false, nil
	}
	return u, err == nil, err
}

func (p *postgresUserStore) GetUserByEmail(ctx context.Context, email string) (store.User, bool, error) {
	ps := (*PostgresStore)(p)
	query := `SELECT id, email, name, created_at, updated_at FROM users WHERE email = $1`
	var u store.User
	err := ps.db.QueryRowContext(ctx, query, email).Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return u, false, nil
	}
	return u, err == nil, err
}

func (p *postgresUserStore) CreateUserOrg(ctx context.Context, uo store.UserOrg) error {
	ps := (*PostgresStore)(p)
	query := `INSERT INTO user_orgs (user_id, org_id, role) VALUES ($1, $2, $3)`
	_, err := ps.db.ExecContext(ctx, query, uo.UserID, uo.OrgID, uo.Role)
	return err
}

func (p *postgresUserStore) ListUserOrgs(ctx context.Context, userID string) ([]store.UserOrg, error) {
	ps := (*PostgresStore)(p)
	query := `SELECT user_id, org_id, role FROM user_orgs WHERE user_id = $1`
	rows, err := ps.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []store.UserOrg
	for rows.Next() {
		var uo store.UserOrg
		if err := rows.Scan(&uo.UserID, &uo.OrgID, &uo.Role); err != nil {
			return nil, err
		}
		result = append(result, uo)
	}
	return result, nil
}

type postgresOrganizationStore PostgresStore

func (p *postgresOrganizationStore) CreateOrganization(ctx context.Context, o *store.Organization) error {
	ps := (*PostgresStore)(p)
	// Let PostgreSQL generate the UUID automatically
	query := `INSERT INTO organizations (name) VALUES ($1) RETURNING id, created_at`
	err := ps.db.QueryRowContext(ctx, query, o.Name).Scan(&o.ID, &o.CreatedAt)
	return err
}

func (p *postgresOrganizationStore) GetOrganization(ctx context.Context, orgID string) (store.Organization, error) {
	ps := (*PostgresStore)(p)
	query := `SELECT id, name, created_at FROM organizations WHERE id = $1`
	var o store.Organization
	err := ps.db.QueryRowContext(ctx, query, orgID).Scan(&o.ID, &o.Name, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return o, fmt.Errorf("organization not found")
	}
	return o, err
}

// Simple ID generator
func generateID() string {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return "fallback"
	}
	return hex.EncodeToString(b[:])
}
