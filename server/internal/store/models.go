package store

import "time"
import "github.com/ziyad/cms-ai/server/internal/auth"

type TemplateStatus string

const (
	TemplateDraft     TemplateStatus = "Draft"
	TemplatePublished TemplateStatus = "Published"
	TemplateArchived  TemplateStatus = "Archived"
)

type Template struct {
	ID              string         `json:"id"`
	OrgID           string         `json:"orgId"`
	OwnerUserID     string         `json:"ownerUserId"`
	Name            string         `json:"name"`
	Status          TemplateStatus `json:"status"`
	CurrentVersion  *string        `json:"currentVersionId"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	LatestVersionNo int            `json:"latestVersionNo"`
}

type TemplateVersion struct {
	ID        string    `json:"id"`
	Template  string    `json:"templateId"`
	OrgID     string    `json:"orgId"`
	VersionNo int       `json:"versionNo"`
	SpecJSON  any       `json:"spec"`
	CreatedBy string    `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
}

type BrandKit struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"orgId"`
	Name      string    `json:"name"`
	Tokens    any       `json:"tokens"`
	CreatedAt time.Time `json:"createdAt"`
}

type AssetType string

const (
	AssetPPTX AssetType = "pptx"
	AssetPNG  AssetType = "png"
	AssetFile AssetType = "file"
)

type Asset struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"orgId"`
	Type      AssetType `json:"type"`
	Path      string    `json:"path"`
	Mime      string    `json:"mime"`
	CreatedAt time.Time `json:"createdAt"`
}

type JobStatus string

type JobType string

const (
	JobQueued     JobStatus = "Queued"
	JobRunning    JobStatus = "Running"
	JobDone       JobStatus = "Done"
	JobFailed     JobStatus = "Failed"
	JobRetry      JobStatus = "Retry"
	JobDeadLetter JobStatus = "DeadLetter"

	JobRender  JobType = "render"
	JobPreview JobType = "preview"
	JobExport  JobType = "export"
)

type Job struct {
	ID              string            `json:"id"`
	OrgID           string            `json:"orgId"`
	Type            JobType           `json:"type"`
	Status          JobStatus         `json:"status"`
	InputRef        string            `json:"inputRef"`
	OutputRef       string            `json:"outputRef,omitempty"`
	Error           string            `json:"error,omitempty"`
	RetryCount      int               `json:"retryCount"`
	MaxRetries      int               `json:"maxRetries"`
	LastRetryAt     *time.Time        `json:"lastRetryAt,omitempty"`
	DeduplicationID string            `json:"deduplicationId,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
}

type MeteringEvent struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"orgId"`
	UserID    string    `json:"userId"`
	Type      string    `json:"eventType"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"createdAt"`
}

type AuditLog struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"orgId"`
	ActorID   string    `json:"actorUserId"`
	Action    string    `json:"action"`
	TargetRef string    `json:"targetRef"`
	Metadata  any       `json:"metadata"`
	CreatedAt time.Time `json:"createdAt"`
}

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UserOrg struct {
	UserID string    `json:"userId"`
	OrgID  string    `json:"orgId"`
	Role   auth.Role `json:"role"`
}
