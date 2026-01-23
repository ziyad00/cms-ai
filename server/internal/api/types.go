package api

type AnalyzeTemplateRequest struct {
	Prompt string `json:"prompt"`
}

type RequiredField struct {
	Key         string   `json:"key"`
	Label       string   `json:"label"`
	Type        string   `json:"type"` // text, number, currency, percentage, date, list
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
	Prompt      string                 `json:"prompt"`
	Name        string                 `json:"name,omitempty"`
	BrandKitID  string                 `json:"brandKitId,omitempty"`
	RTL         bool                   `json:"rtl"`
	Language    string                 `json:"language,omitempty"`
	Tone        string                 `json:"tone,omitempty"`
	ContentData map[string]interface{} `json:"contentData,omitempty"`
}

type CreateTemplateRequest struct {
	Name string `json:"name"`
}

type CreateVersionRequest struct {
	Spec any `json:"spec"`
}

type PatchVersionRequest struct {
	Spec any `json:"spec"`
}

type UsageResponse struct {
	OrgID   string         `json:"orgId"`
	Limits  map[string]int `json:"limits"`
	Used    map[string]int `json:"used"`
	Blocked bool           `json:"blocked"`
}
