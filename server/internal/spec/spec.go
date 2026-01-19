package spec

type TemplateSpec struct {
	Tokens      map[string]any `json:"tokens"`
	Constraints Constraints    `json:"constraints"`
	Layouts     []Layout       `json:"layouts"`
}

type Constraints struct {
	SafeMargin float64 `json:"safeMargin"`
}

type Layout struct {
	Name         string        `json:"name"`
	Placeholders []Placeholder `json:"placeholders"`
}

type Placeholder struct {
	ID       string   `json:"id"`
	Type     string   `json:"type,omitempty"`
	Geometry Geometry `json:"geometry"`
}

type Geometry struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

type ValidationError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}
