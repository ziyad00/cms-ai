package api

type AnalyzeTemplateRequest struct {
	Prompt string `json:"prompt" validate:"required,min=3"`
}

type RequiredField struct {
	Key         string   `json:"key" validate:"required"`
	Label       string   `json:"label" validate:"required"`
	Type        string   `json:"type" validate:"required,oneof=text number currency percentage date list"`
	Required    bool     `json:"required"`
	Example     string   `json:"example"`
	Options     []string `json:"options,omitempty"` // for select fields
	Description string   `json:"description,omitempty"`
}

type AnalyzeTemplateResponse struct {
	TemplateType    string          `json:"templateType"`
	SuggestedName   string          `json:"suggestedName"`
	RequiredFields  []RequiredField `json:"requiredFields"`
	EstimatedSlides int             `json:"estimatedSlides"`
	Description     string          `json:"description"`
}

type GenerateTemplateRequest struct {
	Prompt      string                 `json:"prompt" validate:"required,min=10"`
	Name        string                 `json:"name,omitempty"`
	BrandKitID  string                 `json:"brandKitId,omitempty"`
	RTL         bool                   `json:"rtl"`
	Language    string                 `json:"language,omitempty"`
	Tone        string                 `json:"tone,omitempty"`
	ContentData map[string]interface{} `json:"contentData,omitempty"`
}

type CreateTemplateRequest struct {
	Name string `json:"name" validate:"required,min=3"`
}

type SlideOutline struct {
	SlideNumber int      `json:"slide_number" validate:"required"`
	Title       string   `json:"title" validate:"required"`
	Content     []string `json:"content"`
}

type DeckOutline struct {
	Slides []SlideOutline `json:"slides" validate:"required,dive"`
}

type CreateDeckOutlineRequest struct {
	Prompt  string `json:"prompt" validate:"required,min=5"`
	Content string `json:"content" validate:"required,min=10"`
}

type CreateDeckOutlineResponse struct {
	Outline DeckOutline `json:"outline"`
}

type CreateDeckRequest struct {
	Name                  string `json:"name" validate:"required,min=3"`
	SourceTemplateVersion string `json:"sourceTemplateVersionId" validate:"required"`
	Content               string `json:"content" validate:"required,min=10"`
	Outline               any    `json:"outline,omitempty"`
}

type CreateDeckVersionRequest struct {
	Spec any `json:"spec" validate:"required"`
}

type CreateVersionRequest struct {
	Spec any `json:"spec" validate:"required"`
}

type PatchVersionRequest struct {
	Spec any `json:"spec" validate:"required"`
}

type UsageResponse struct {
	OrgID   string         `json:"orgId"`
	Limits  map[string]int `json:"limits"`
	Used    map[string]int `json:"used"`
	Blocked bool           `json:"blocked"`
}
