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
	ID              string         `json:"id" gorm:"primaryKey"`
	OrgID           string         `json:"orgId" gorm:"index"`
	OwnerUserID     string         `json:"ownerUserId" gorm:"index"`
	Name            string         `json:"name"`
	Status          TemplateStatus `json:"status"`
	CurrentVersion  *string        `json:"currentVersionId" gorm:"index"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	LatestVersionNo int            `json:"latestVersionNo"`
}

type Deck struct {
	ID                    string    `json:"id" gorm:"primaryKey"`
	OrgID                 string    `json:"orgId" gorm:"index"`
	OwnerUserID           string    `json:"ownerUserId" gorm:"index"`
	Name                  string    `json:"name"`
	SourceTemplateVersion string    `json:"sourceTemplateVersionId" gorm:"index"`
	CurrentVersion        *string   `json:"currentVersionId" gorm:"index"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
	LatestVersionNo       int       `json:"latestVersionNo"`
	Content               string    `json:"content"`
}

type DeckVersion struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Deck      string    `json:"deckId" gorm:"index"`
	OrgID     string    `json:"orgId" gorm:"index"`
	VersionNo int       `json:"versionNo"`
	SpecJSON  any       `json:"spec" gorm:"type:jsonb"`
	CreatedBy string    `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
}

type TemplateVersion struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Template  string    `json:"templateId" gorm:"index"`
	OrgID     string    `json:"orgId" gorm:"index"`
	VersionNo int       `json:"versionNo"`
	SpecJSON  any       `json:"spec" gorm:"type:jsonb"`
	CreatedBy string    `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
}

type BrandKit struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	OrgID     string    `json:"orgId" gorm:"index"`
	Name      string    `json:"name"`
	Tokens    any       `json:"tokens" gorm:"type:jsonb"`
	CreatedAt time.Time `json:"createdAt"`
}

type AssetType string

const (
	AssetPPTX AssetType = "pptx"
	AssetPNG  AssetType = "png"
	AssetFile AssetType = "file"
)

type Asset struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	OrgID     string    `json:"orgId" gorm:"index"`
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
	JobGenerate JobType = "generate"
	JobBind     JobType = "bind"
)

type Job struct {
	ID              string            `json:"id" gorm:"primaryKey"`
	OrgID           string            `json:"orgId" gorm:"index"`
	Type            JobType           `json:"type" gorm:"index"`
	Status          JobStatus         `json:"status" gorm:"index"`
	InputRef        string            `json:"inputRef" gorm:"index"`
	OutputRef       string            `json:"outputRef,omitempty"`
	Error           string            `json:"error,omitempty"`
	RetryCount      int               `json:"retryCount"`
	MaxRetries      int               `json:"maxRetries"`
	LastRetryAt     *time.Time        `json:"lastRetryAt,omitempty"`
	DeduplicationID string            `json:"deduplicationId,omitempty" gorm:"index"`
	Metadata        *map[string]string `json:"metadata,omitempty" gorm:"type:jsonb"`
	ProgressStep    string            `json:"progressStep,omitempty"`
	ProgressPct     int               `json:"progressPct,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
}

type MeteringEvent struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	OrgID     string    `json:"orgId" gorm:"index"`
	UserID    string    `json:"userId" gorm:"index"`
	Type      string    `json:"eventType" gorm:"index"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"createdAt"`
}

type AuditLog struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	OrgID     string    `json:"orgId" gorm:"index"`
	ActorID   string    `json:"actorUserId" gorm:"index"`
	Action    string    `json:"action" gorm:"index"`
	TargetRef string    `json:"targetRef" gorm:"index"`
	Metadata  any       `json:"metadata" gorm:"type:jsonb"`
	CreatedAt time.Time `json:"createdAt"`
}

type User struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"uniqueIndex"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Organization struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UserOrg struct {
	UserID string    `json:"userId" gorm:"primaryKey"`
	OrgID  string    `json:"orgId" gorm:"primaryKey"`
	Role   auth.Role `json:"role"`
}
