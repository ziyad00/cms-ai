package store

import "context"

type Store interface {
	Templates() TemplateStore
	Decks() DeckStore
	BrandKits() BrandKitStore
	Assets() AssetStore
	Jobs() JobStore
	Metering() MeteringStore
	Audit() AuditStore
	Users() UserStore
	Organizations() OrganizationStore
}

type DeckStore interface {
	CreateDeck(ctx context.Context, d Deck) (Deck, error)
	ListDecks(ctx context.Context, orgID string) ([]Deck, error)
	GetDeck(ctx context.Context, orgID, id string) (Deck, bool, error)
	UpdateDeck(ctx context.Context, d Deck) (Deck, error)

	CreateDeckVersion(ctx context.Context, v DeckVersion) (DeckVersion, error)
	ListDeckVersions(ctx context.Context, orgID, deckID string) ([]DeckVersion, error)
	GetDeckVersion(ctx context.Context, orgID, versionID string) (DeckVersion, bool, error)
}

type AssetStore interface {
	Create(ctx context.Context, a Asset) (Asset, error)
	Get(ctx context.Context, orgID, id string) (Asset, bool, error)
	Store(ctx context.Context, orgID, assetID string, data []byte) (string, error)
}

type TemplateStore interface {
	CreateTemplate(ctx context.Context, t Template) (Template, error)
	ListTemplates(ctx context.Context, orgID string) ([]Template, error)
	GetTemplate(ctx context.Context, orgID, id string) (Template, bool, error)
	UpdateTemplate(ctx context.Context, t Template) (Template, error)

	CreateVersion(ctx context.Context, v TemplateVersion) (TemplateVersion, error)
	ListVersions(ctx context.Context, orgID, templateID string) ([]TemplateVersion, error)
	GetVersion(ctx context.Context, orgID, versionID string) (TemplateVersion, bool, error)
}

type BrandKitStore interface {
	Create(ctx context.Context, b BrandKit) (BrandKit, error)
	List(ctx context.Context, orgID string) ([]BrandKit, error)
}

type JobStore interface {
	Enqueue(ctx context.Context, j Job) (Job, error)
	EnqueueWithDeduplication(ctx context.Context, j Job) (Job, bool, error)
	Get(ctx context.Context, orgID, jobID string) (Job, bool, error)
	GetByDeduplicationID(ctx context.Context, orgID, dedupID string) (Job, bool, error)
	Update(ctx context.Context, j Job) (Job, error)
	ListQueued(ctx context.Context) ([]Job, error)
	ListRetry(ctx context.Context) ([]Job, error)
	ListDeadLetter(ctx context.Context) ([]Job, error)
	ListByInputRef(ctx context.Context, orgID, inputRef string, jobType JobType) ([]Job, error)
	MoveToDeadLetter(ctx context.Context, jobID string) error
	RetryDeadLetterJob(ctx context.Context, jobID string) error
}

type MeteringStore interface {
	Record(ctx context.Context, e MeteringEvent) (MeteringEvent, error)
	SumByType(ctx context.Context, orgID string, eventType string) (int, error)
}

type AuditStore interface {
	Append(ctx context.Context, a AuditLog) (AuditLog, error)
}

type UserStore interface {
	CreateUser(ctx context.Context, u *User) error
	GetUser(ctx context.Context, userID string) (User, bool, error)
	GetUserByEmail(ctx context.Context, email string) (User, bool, error)
	CreateUserOrg(ctx context.Context, uo UserOrg) error
	ListUserOrgs(ctx context.Context, userID string) ([]UserOrg, error)
}

type OrganizationStore interface {
	CreateOrganization(ctx context.Context, o *Organization) error
	GetOrganization(ctx context.Context, orgID string) (Organization, error)
}
