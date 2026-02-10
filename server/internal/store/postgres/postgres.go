package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/ziyad/cms-ai/server/internal/store"
)

type PostgresStore struct {
	db *gorm.DB
}

func New(dsn string) (*PostgresStore, error) {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	}
	if os.Getenv("DEV_MODE") == "true" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}

	// Idempotent Migrator: Safely transition legacy SQL constraints to GORM management.
	// We use the Migrator API to check for and remove legacy names so AutoMigrate can 
	// establish its own naming convention without conflicts.
	m := db.Migrator()
	if m.HasConstraint(&store.User{}, "users_email_key") {
		log.Printf("🔄 GORM: Migrating legacy constraint 'users_email_key' to GORM convention...")
		_ = m.DropConstraint(&store.User{}, "users_email_key")
	}

	// Auto-migrate all models to ensure schema is always in sync
	log.Printf("🚀 GORM: Running auto-migration...")
	err = db.AutoMigrate(
		&store.Organization{},
		&store.User{},
		&store.UserOrg{},
		&store.Template{},
		&store.TemplateVersion{},
		&store.Deck{},
		&store.DeckVersion{},
		&store.BrandKit{},
		&store.Asset{},
		&store.Job{},
		&store.MeteringEvent{},
		&store.AuditLog{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate: %w", err)
	}

	return &PostgresStore{db: db}, nil
}

func (p *PostgresStore) Close() error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// DB exposes the underlying database connection for diagnostics
func (p *PostgresStore) DB() (*sql.DB, error) {
	return p.db.DB()
}

func (p *PostgresStore) Templates() store.TemplateStore         { return (*postgresTemplateStore)(p) }
func (p *PostgresStore) Decks() store.DeckStore                 { return (*postgresDeckStore)(p) }
func (p *PostgresStore) BrandKits() store.BrandKitStore         { return (*postgresBrandKitStore)(p) }
func (p *PostgresStore) Assets() store.AssetStore               { return (*postgresAssetStore)(p) }
func (p *PostgresStore) Jobs() store.JobStore                   { return (*postgresJobStore)(p) }
func (p *PostgresStore) Metering() store.MeteringStore         { return (*postgresMeteringStore)(p) }
func (p *PostgresStore) Audit() store.AuditStore               { return (*postgresAuditStore)(p) }
func (p *PostgresStore) Users() store.UserStore                 { return (*postgresUserStore)(p) }
func (p *PostgresStore) Organizations() store.OrganizationStore { return (*postgresOrganizationStore)(p) }

type postgresTemplateStore PostgresStore

func (p *postgresTemplateStore) CreateTemplate(ctx context.Context, t store.Template) (store.Template, error) {
	ps := (*PostgresStore)(p)
	if t.ID == "" {
		t.ID = newID("tmpl")
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	t.UpdatedAt = t.CreatedAt
	err := ps.db.WithContext(ctx).Create(&t).Error
	return t, err
}

func (p *postgresTemplateStore) ListTemplates(ctx context.Context, orgID string) ([]store.Template, error) {
	ps := (*PostgresStore)(p)
	var ts []store.Template
	err := ps.db.WithContext(ctx).Where("org_id = ?", orgID).Find(&ts).Error
	return ts, err
}

func (p *postgresTemplateStore) GetTemplate(ctx context.Context, orgID, id string) (store.Template, bool, error) {
	ps := (*PostgresStore)(p)
	var t store.Template
	err := ps.db.WithContext(ctx).Where("org_id = ? AND id = ?", orgID, id).First(&t).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return store.Template{}, false, nil
		}
		return store.Template{}, false, err
	}
	return t, true, nil
}

func (p *postgresTemplateStore) UpdateTemplate(ctx context.Context, t store.Template) (store.Template, error) {
	ps := (*PostgresStore)(p)
	t.UpdatedAt = time.Now().UTC()
	err := ps.db.WithContext(ctx).Save(&t).Error
	return t, err
}

func (p *postgresTemplateStore) CreateVersion(ctx context.Context, v store.TemplateVersion) (store.TemplateVersion, error) {
	ps := (*PostgresStore)(p)
	if v.ID == "" {
		v.ID = newID("tv")
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now().UTC()
	}
	err := ps.db.WithContext(ctx).Create(&v).Error
	return v, err
}

func (p *postgresTemplateStore) ListVersions(ctx context.Context, orgID, templateID string) ([]store.TemplateVersion, error) {
	ps := (*PostgresStore)(p)
	var vs []store.TemplateVersion
	err := ps.db.WithContext(ctx).Where("org_id = ? AND template_id = ?", orgID, templateID).Order("version_no DESC").Find(&vs).Error
	return vs, err
}

func (p *postgresTemplateStore) GetVersion(ctx context.Context, orgID, versionID string) (store.TemplateVersion, bool, error) {
	ps := (*PostgresStore)(p)
	var v store.TemplateVersion
	err := ps.db.WithContext(ctx).Where("org_id = ? AND id = ?", orgID, versionID).First(&v).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return store.TemplateVersion{}, false, nil
		}
		return store.TemplateVersion{}, false, err
	}
	return v, true, nil
}

type postgresDeckStore PostgresStore

func (p *postgresDeckStore) CreateDeck(ctx context.Context, d store.Deck) (store.Deck, error) {
	ps := (*PostgresStore)(p)
	if d.ID == "" {
		d.ID = newID("deck")
	}
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now().UTC()
	}
	d.UpdatedAt = d.CreatedAt
	err := ps.db.WithContext(ctx).Create(&d).Error
	return d, err
}

func (p *postgresDeckStore) ListDecks(ctx context.Context, orgID string) ([]store.Deck, error) {
	ps := (*PostgresStore)(p)
	var ds []store.Deck
	err := ps.db.WithContext(ctx).Where("org_id = ?", orgID).Order("updated_at DESC").Find(&ds).Error
	return ds, err
}

func (p *postgresDeckStore) GetDeck(ctx context.Context, orgID, id string) (store.Deck, bool, error) {
	ps := (*PostgresStore)(p)
	var d store.Deck
	err := ps.db.WithContext(ctx).Where("org_id = ? AND id = ?", orgID, id).First(&d).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return store.Deck{}, false, nil
		}
		return store.Deck{}, false, err
	}
	return d, true, nil
}

func (p *postgresDeckStore) UpdateDeck(ctx context.Context, d store.Deck) (store.Deck, error) {
	ps := (*PostgresStore)(p)
	d.UpdatedAt = time.Now().UTC()
	err := ps.db.WithContext(ctx).Save(&d).Error
	return d, err
}

func (p *postgresDeckStore) CreateDeckVersion(ctx context.Context, v store.DeckVersion) (store.DeckVersion, error) {
	ps := (*PostgresStore)(p)
	if v.ID == "" {
		v.ID = newID("dv")
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now().UTC()
	}
	err := ps.db.WithContext(ctx).Create(&v).Error
	return v, err
}

func (p *postgresDeckStore) ListDeckVersions(ctx context.Context, orgID, deckID string) ([]store.DeckVersion, error) {
	ps := (*PostgresStore)(p)
	var vs []store.DeckVersion
	err := ps.db.WithContext(ctx).Where("org_id = ? AND deck_id = ?", orgID, deckID).Order("version_no DESC").Find(&vs).Error
	return vs, err
}

func (p *postgresDeckStore) GetDeckVersion(ctx context.Context, orgID, versionID string) (store.DeckVersion, bool, error) {
	ps := (*PostgresStore)(p)
	var v store.DeckVersion
	err := ps.db.WithContext(ctx).Where("org_id = ? AND id = ?", orgID, versionID).First(&v).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return store.DeckVersion{}, false, nil
		}
		return store.DeckVersion{}, false, err
	}
	return v, true, nil
}

type postgresBrandKitStore PostgresStore

func (p *postgresBrandKitStore) Create(ctx context.Context, b store.BrandKit) (store.BrandKit, error) {
	ps := (*PostgresStore)(p)
	if b.ID == "" {
		b.ID = newID("bk")
	}
	b.CreatedAt = time.Now().UTC()
	err := ps.db.WithContext(ctx).Create(&b).Error
	return b, err
}

func (p *postgresBrandKitStore) List(ctx context.Context, orgID string) ([]store.BrandKit, error) {
	ps := (*PostgresStore)(p)
	var bks []store.BrandKit
	err := ps.db.WithContext(ctx).Where("org_id = ?", orgID).Find(&bks).Error
	return bks, err
}

type postgresAssetStore PostgresStore

func (p *postgresAssetStore) Create(ctx context.Context, a store.Asset) (store.Asset, error) {
	ps := (*PostgresStore)(p)
	if a.ID == "" {
		a.ID = newID("asset")
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	err := ps.db.WithContext(ctx).Create(&a).Error
	return a, err
}

func (p *postgresAssetStore) Get(ctx context.Context, orgID, id string) (store.Asset, bool, error) {
	ps := (*PostgresStore)(p)
	var a store.Asset
	err := ps.db.WithContext(ctx).Where("org_id = ? AND id = ?", orgID, id).First(&a).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return store.Asset{}, false, nil
		}
		return store.Asset{}, false, err
	}
	return a, true, nil
}

type postgresJobStore PostgresStore

func (p *postgresJobStore) Enqueue(ctx context.Context, j store.Job) (store.Job, error) {
	ps := (*PostgresStore)(p)
	if j.ID == "" {
		j.ID = newID("job")
	}
	if j.MaxRetries == 0 {
		j.MaxRetries = 3
	}
	j.CreatedAt = time.Now().UTC()
	j.UpdatedAt = j.CreatedAt
	err := ps.db.WithContext(ctx).Create(&j).Error
	return j, err
}

func (p *postgresJobStore) EnqueueWithDeduplication(ctx context.Context, j store.Job) (store.Job, bool, error) {
	ps := (*PostgresStore)(p)
	if j.DeduplicationID != "" {
		var existingJob store.Job
		err := ps.db.WithContext(ctx).Where("org_id = ? AND deduplication_id = ?", j.OrgID, j.DeduplicationID).Order("created_at DESC").First(&existingJob).Error
		if err == nil {
			if existingJob.Status == store.JobQueued || existingJob.Status == store.JobRunning || existingJob.Status == store.JobRetry || existingJob.Status == store.JobDone {
				return existingJob, true, nil
			}
		}
	}
	inserted, err := p.Enqueue(ctx, j)
	return inserted, false, err
}

func (p *postgresJobStore) Get(ctx context.Context, orgID, jobID string) (store.Job, bool, error) {
	ps := (*PostgresStore)(p)
	var j store.Job
	err := ps.db.WithContext(ctx).Where("org_id = ? AND id = ?", orgID, jobID).First(&j).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return store.Job{}, false, nil
		}
		return store.Job{}, false, err
	}
	return j, true, nil
}

func (p *postgresJobStore) GetByDeduplicationID(ctx context.Context, orgID, dedupID string) (store.Job, bool, error) {
	ps := (*PostgresStore)(p)
	var j store.Job
	err := ps.db.WithContext(ctx).Where("org_id = ? AND deduplication_id = ?", orgID, dedupID).Order("created_at DESC").First(&j).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return store.Job{}, false, nil
		}
		return store.Job{}, false, err
	}
	return j, true, nil
}

func (p *postgresJobStore) Update(ctx context.Context, j store.Job) (store.Job, error) {
	ps := (*PostgresStore)(p)
	j.UpdatedAt = time.Now().UTC()
	err := ps.db.WithContext(ctx).Save(&j).Error
	return j, err
}

func (p *postgresJobStore) ListQueued(ctx context.Context) ([]store.Job, error) {
	ps := (*PostgresStore)(p)
	var jobs []store.Job
	err := ps.db.WithContext(ctx).Where("status = ?", store.JobQueued).Order("created_at ASC").Find(&jobs).Error
	return jobs, err
}

func (p *postgresJobStore) ListRetry(ctx context.Context) ([]store.Job, error) {
	ps := (*PostgresStore)(p)
	var jobs []store.Job
	err := ps.db.WithContext(ctx).Where("status = ?", store.JobRetry).Order("last_retry_at ASC").Find(&jobs).Error
	return jobs, err
}

func (p *postgresJobStore) ListDeadLetter(ctx context.Context) ([]store.Job, error) {
	ps := (*PostgresStore)(p)
	var jobs []store.Job
	err := ps.db.WithContext(ctx).Where("status = ?", store.JobDeadLetter).Order("updated_at DESC").Find(&jobs).Error
	return jobs, err
}

func (p *postgresJobStore) ListByInputRef(ctx context.Context, orgID, inputRef string, jobType store.JobType) ([]store.Job, error) {
	ps := (*PostgresStore)(p)
	var jobs []store.Job
	err := ps.db.WithContext(ctx).Where("org_id = ? AND input_ref = ? AND type = ?", orgID, inputRef, jobType).Order("updated_at DESC").Find(&jobs).Error
	return jobs, err
}

func (p *postgresJobStore) MoveToDeadLetter(ctx context.Context, jobID string) error {
	ps := (*PostgresStore)(p)
	return ps.db.WithContext(ctx).Model(&store.Job{}).Where("id = ?", jobID).Update("status", store.JobDeadLetter).Error
}

func (p *postgresJobStore) RetryDeadLetterJob(ctx context.Context, jobID string) error {
	ps := (*PostgresStore)(p)
	return ps.db.WithContext(ctx).Model(&store.Job{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":      store.JobQueued,
		"retry_count": 0,
		"error":       "",
	}).Error
}

type postgresMeteringStore PostgresStore

func (p *postgresMeteringStore) Record(ctx context.Context, e store.MeteringEvent) (store.MeteringEvent, error) {
	ps := (*PostgresStore)(p)
	if e.ID == "" {
		e.ID = newID("met")
	}
	e.CreatedAt = time.Now().UTC()
	err := ps.db.WithContext(ctx).Create(&e).Error
	return e, err
}

func (p *postgresMeteringStore) SumByType(ctx context.Context, orgID string, eventType string) (int, error) {
	ps := (*PostgresStore)(p)
	var sum int64
	err := ps.db.WithContext(ctx).Model(&store.MeteringEvent{}).Where("org_id = ? AND event_type = ?", orgID, eventType).Select("SUM(quantity)").Scan(&sum).Error
	return int(sum), err
}

type postgresAuditStore PostgresStore

func (p *postgresAuditStore) Append(ctx context.Context, a store.AuditLog) (store.AuditLog, error) {
	ps := (*PostgresStore)(p)
	if a.ID == "" {
		a.ID = newID("aud")
	}
	a.CreatedAt = time.Now().UTC()
	err := ps.db.WithContext(ctx).Create(&a).Error
	return a, err
}

type postgresUserStore PostgresStore

func (p *postgresUserStore) CreateUser(ctx context.Context, u *store.User) error {
	ps := (*PostgresStore)(p)
	if u.ID == "" {
		u.ID = newID("user")
	}
	u.CreatedAt = time.Now().UTC()
	u.UpdatedAt = u.CreatedAt
	return ps.db.WithContext(ctx).Create(u).Error
}

func (p *postgresUserStore) GetUser(ctx context.Context, userID string) (store.User, bool, error) {
	ps := (*PostgresStore)(p)
	var u store.User
	err := ps.db.WithContext(ctx).Where("id = ?", userID).First(&u).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return store.User{}, false, nil
		}
		return store.User{}, false, err
	}
	return u, true, nil
}

func (p *postgresUserStore) GetUserByEmail(ctx context.Context, email string) (store.User, bool, error) {
	ps := (*PostgresStore)(p)
	var u store.User
	err := ps.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return store.User{}, false, nil
		}
		return store.User{}, false, err
	}
	return u, true, nil
}

func (p *postgresUserStore) CreateUserOrg(ctx context.Context, uo store.UserOrg) error {
	ps := (*PostgresStore)(p)
	return ps.db.WithContext(ctx).Create(&uo).Error
}

func (p *postgresUserStore) ListUserOrgs(ctx context.Context, userID string) ([]store.UserOrg, error) {
	ps := (*PostgresStore)(p)
	var uos []store.UserOrg
	err := ps.db.WithContext(ctx).Where("user_id = ?", userID).Find(&uos).Error
	return uos, err
}

type postgresOrganizationStore PostgresStore

func (p *postgresOrganizationStore) CreateOrganization(ctx context.Context, o *store.Organization) error {
	ps := (*PostgresStore)(p)
	if o.ID == "" {
		o.ID = newID("org")
	}
	o.CreatedAt = time.Now().UTC()
	o.UpdatedAt = o.CreatedAt
	return ps.db.WithContext(ctx).Create(o).Error
}

func (p *postgresOrganizationStore) GetOrganization(ctx context.Context, orgID string) (store.Organization, error) {
	ps := (*PostgresStore)(p)
	var o store.Organization
	err := ps.db.WithContext(ctx).Where("id = ?", orgID).First(&o).Error
	return o, err
}

func newID(prefix string) string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return prefix + "-" + hex.EncodeToString(b[:])
}