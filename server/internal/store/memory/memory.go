package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ziyad/cms-ai/server/internal/store"
)

type MemoryStore struct {
	mu sync.Mutex

	templates map[string]store.Template
	versions  map[string]store.TemplateVersion
	brandKits map[string]store.BrandKit
	assets    map[string]store.Asset
	assetData map[string][]byte
	jobs      map[string]store.Job
	metering  []store.MeteringEvent
	audit     []store.AuditLog
	users     map[string]store.User
	orgs      map[string]store.Organization
	userOrgs  []store.UserOrg
}

func New() *MemoryStore {
	return &MemoryStore{
		templates: map[string]store.Template{},
		versions:  map[string]store.TemplateVersion{},
		brandKits: map[string]store.BrandKit{},
		assets:    map[string]store.Asset{},
		assetData: map[string][]byte{},
		jobs:      map[string]store.Job{},
		metering:  []store.MeteringEvent{},
		audit:     []store.AuditLog{},
		users:     map[string]store.User{},
		orgs:      map[string]store.Organization{},
		userOrgs:  []store.UserOrg{},
	}
}

func (m *MemoryStore) Templates() store.TemplateStore         { return (*templateStore)(m) }
func (m *MemoryStore) BrandKits() store.BrandKitStore         { return (*brandKitStore)(m) }
func (m *MemoryStore) Assets() store.AssetStore               { return (*assetStore)(m) }
func (m *MemoryStore) Jobs() store.JobStore                   { return (*jobStore)(m) }
func (m *MemoryStore) Metering() store.MeteringStore          { return (*meteringStore)(m) }
func (m *MemoryStore) Audit() store.AuditStore                { return (*auditStore)(m) }
func (m *MemoryStore) Users() store.UserStore                 { return (*userStore)(m) }
func (m *MemoryStore) Organizations() store.OrganizationStore { return (*organizationStore)(m) }

type templateStore MemoryStore

type brandKitStore MemoryStore

type assetStore MemoryStore

type jobStore MemoryStore

type meteringStore MemoryStore

type auditStore MemoryStore

type userStore MemoryStore

type organizationStore MemoryStore

var errNotFound = errors.New("not found")

func (m *templateStore) CreateTemplate(_ context.Context, t store.Template) (store.Template, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	now := time.Now().UTC()
	t.CreatedAt = now
	t.UpdatedAt = now
	ms.templates[t.ID] = t
	return t, nil
}

func (m *templateStore) ListTemplates(_ context.Context, orgID string) ([]store.Template, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	out := make([]store.Template, 0, len(ms.templates))
	for _, t := range ms.templates {
		if t.OrgID == orgID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (m *templateStore) GetTemplate(_ context.Context, orgID, id string) (store.Template, bool, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	t, ok := ms.templates[id]
	if !ok || t.OrgID != orgID {
		return store.Template{}, false, nil
	}
	return t, true, nil
}

func (m *templateStore) UpdateTemplate(_ context.Context, t store.Template) (store.Template, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, ok := ms.templates[t.ID]; !ok {
		return store.Template{}, errNotFound
	}
	t.UpdatedAt = time.Now().UTC()
	ms.templates[t.ID] = t
	return t, nil
}

func (m *templateStore) CreateVersion(_ context.Context, v store.TemplateVersion) (store.TemplateVersion, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	v.CreatedAt = time.Now().UTC()
	ms.versions[v.ID] = v
	return v, nil
}

func (m *templateStore) ListVersions(_ context.Context, orgID, templateID string) ([]store.TemplateVersion, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	out := []store.TemplateVersion{}
	for _, v := range ms.versions {
		if v.OrgID == orgID && v.Template == templateID {
			out = append(out, v)
		}
	}
	return out, nil
}

func (m *templateStore) GetVersion(_ context.Context, orgID, versionID string) (store.TemplateVersion, bool, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	v, ok := ms.versions[versionID]
	if !ok || v.OrgID != orgID {
		return store.TemplateVersion{}, false, nil
	}
	return v, true, nil
}

func (m *brandKitStore) Create(_ context.Context, b store.BrandKit) (store.BrandKit, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	b.CreatedAt = time.Now().UTC()
	ms.brandKits[b.ID] = b
	return b, nil
}

func (m *brandKitStore) List(_ context.Context, orgID string) ([]store.BrandKit, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	out := []store.BrandKit{}
	for _, b := range ms.brandKits {
		if b.OrgID == orgID {
			out = append(out, b)
		}
	}
	return out, nil
}

func (m *assetStore) Create(_ context.Context, a store.Asset) (store.Asset, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	a.CreatedAt = time.Now().UTC()
	ms.assets[a.ID] = a
	return a, nil
}

func (m *assetStore) Get(_ context.Context, orgID, id string) (store.Asset, bool, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	a, ok := ms.assets[id]
	if !ok || a.OrgID != orgID {
		return store.Asset{}, false, nil
	}
	return a, true, nil
}

func (m *assetStore) Store(_ context.Context, orgID, assetID string, data []byte) (string, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Store data in memory map (in real implementation, use file system or object storage)
	if ms.assetData == nil {
		ms.assetData = make(map[string][]byte)
	}
	ms.assetData[assetID] = data

	// Return a fake path for now
	return fmt.Sprintf("assets/%s/%s", orgID, assetID), nil
}

func (m *jobStore) Enqueue(_ context.Context, j store.Job) (store.Job, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	now := time.Now().UTC()
	j.CreatedAt = now
	j.UpdatedAt = now
	ms.jobs[j.ID] = j
	return j, nil
}

func (m *jobStore) EnqueueWithDeduplication(_ context.Context, j store.Job) (store.Job, bool, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if j.DeduplicationID != "" {
		for _, existingJob := range ms.jobs {
			if existingJob.OrgID == j.OrgID &&
				existingJob.DeduplicationID == j.DeduplicationID &&
				(existingJob.Status == store.JobQueued || existingJob.Status == store.JobRunning) {
				return existingJob, true, nil
			}
		}
	}

	now := time.Now().UTC()
	j.CreatedAt = now
	j.UpdatedAt = now
	ms.jobs[j.ID] = j
	return j, false, nil
}

func (m *jobStore) Get(_ context.Context, orgID, jobID string) (store.Job, bool, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	j, ok := ms.jobs[jobID]
	if !ok || j.OrgID != orgID {
		return store.Job{}, false, nil
	}
	return j, true, nil
}

func (m *jobStore) GetByDeduplicationID(_ context.Context, orgID, dedupID string) (store.Job, bool, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, job := range ms.jobs {
		if job.OrgID == orgID && job.DeduplicationID == dedupID {
			return job, true, nil
		}
	}
	return store.Job{}, false, nil
}

func (m *jobStore) Update(_ context.Context, j store.Job) (store.Job, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, ok := ms.jobs[j.ID]; !ok {
		return store.Job{}, errors.New("not found")
	}
	j.UpdatedAt = time.Now().UTC()
	ms.jobs[j.ID] = j
	return j, nil
}

func (m *jobStore) ListQueued(_ context.Context) ([]store.Job, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	var queued []store.Job
	for _, job := range ms.jobs {
		if job.Status == store.JobQueued {
			queued = append(queued, job)
		}
	}
	return queued, nil
}

func (m *jobStore) ListRetry(_ context.Context) ([]store.Job, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	var retry []store.Job
	for _, job := range ms.jobs {
		if job.Status == store.JobRetry {
			retry = append(retry, job)
		}
	}
	return retry, nil
}

func (m *jobStore) ListDeadLetter(_ context.Context) ([]store.Job, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	var deadLetter []store.Job
	for _, job := range ms.jobs {
		if job.Status == store.JobDeadLetter {
			deadLetter = append(deadLetter, job)
		}
	}
	return deadLetter, nil
}

func (m *jobStore) MoveToDeadLetter(_ context.Context, jobID string) error {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	job, ok := ms.jobs[jobID]
	if !ok {
		return errors.New("job not found")
	}
	job.Status = store.JobDeadLetter
	job.UpdatedAt = time.Now().UTC()
	ms.jobs[jobID] = job
	return nil
}

func (m *jobStore) RetryDeadLetterJob(_ context.Context, jobID string) error {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	job, ok := ms.jobs[jobID]
	if !ok {
		return errors.New("job not found")
	}
	job.Status = store.JobQueued
	job.RetryCount = 0
	job.Error = ""
	job.UpdatedAt = time.Now().UTC()
	ms.jobs[jobID] = job
	return nil
}

func (m *meteringStore) Record(_ context.Context, e store.MeteringEvent) (store.MeteringEvent, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	e.CreatedAt = time.Now().UTC()
	ms.metering = append(ms.metering, e)
	return e, nil
}

func (m *meteringStore) SumByType(_ context.Context, orgID string, eventType string) (int, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	sum := 0
	for _, e := range ms.metering {
		if e.OrgID == orgID && e.Type == eventType {
			sum += e.Quantity
		}
	}
	return sum, nil
}

func (m *auditStore) Append(_ context.Context, a store.AuditLog) (store.AuditLog, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	a.CreatedAt = time.Now().UTC()
	ms.audit = append(ms.audit, a)
	return a, nil
}

func (m *userStore) CreateUser(_ context.Context, u *store.User) error {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	now := time.Now().UTC()
	u.CreatedAt = now
	u.UpdatedAt = now
	ms.users[u.ID] = *u
	return nil
}

func (m *userStore) GetUser(_ context.Context, userID string) (store.User, bool, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	user, ok := ms.users[userID]
	return user, ok, nil
}

func (m *userStore) GetUserByEmail(_ context.Context, email string) (store.User, bool, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, user := range ms.users {
		if user.Email == email {
			return user, true, nil
		}
	}
	return store.User{}, false, nil
}

func (m *userStore) CreateUserOrg(_ context.Context, uo store.UserOrg) error {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.userOrgs = append(ms.userOrgs, uo)
	return nil
}

func (m *userStore) ListUserOrgs(_ context.Context, userID string) ([]store.UserOrg, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	var result []store.UserOrg
	for _, uo := range ms.userOrgs {
		if uo.UserID == userID {
			result = append(result, uo)
		}
	}
	return result, nil
}

func (m *organizationStore) CreateOrganization(_ context.Context, o *store.Organization) error {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	now := time.Now().UTC()
	o.CreatedAt = now
	o.UpdatedAt = now
	ms.orgs[o.ID] = *o
	return nil
}

func (m *organizationStore) GetOrganization(_ context.Context, orgID string) (store.Organization, error) {
	ms := (*MemoryStore)(m)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	org, ok := ms.orgs[orgID]
	if !ok {
		return store.Organization{}, errNotFound
	}
	return org, nil
}
