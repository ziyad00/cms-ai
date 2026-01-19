package api

type GenerateTemplateRequest struct {
	Prompt     string `json:"prompt"`
	Name       string `json:"name,omitempty"`
	BrandKitID string `json:"brandKitId,omitempty"`
	RTL        bool   `json:"rtl"`
	Language   string `json:"language,omitempty"`
	Tone       string `json:"tone,omitempty"`
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
