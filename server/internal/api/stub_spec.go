package api

import (
	"encoding/base64"
	"encoding/json"
)

func assetsSpecBytes(v any) ([]byte, error) {
	// TemplateVersion.SpecJSON is stored as JSONB in Postgres.
	// Depending on the driver/scan target, it can arrive as:
	// - []byte (raw JSON)
	// - string (base64 or JSON string)
	// - map[string]any (already decoded)
	// We need bytes of JSON for parsing.
	switch t := v.(type) {
	case []byte:
		return t, nil
	case json.RawMessage:
		return []byte(t), nil
	case string:
		// If it's already JSON, return bytes; otherwise, it may be base64.
		b := []byte(t)
		if len(b) > 0 && (b[0] == '{' || b[0] == '[') {
			return b, nil
		}
		// Try base64 decode
		decoded, err := base64.StdEncoding.DecodeString(t)
		if err != nil {
			return nil, err
		}
		return decoded, nil
	default:
		return json.Marshal(v)
	}
}

func stubTemplateSpec() map[string]any {
	return map[string]any{
		"tokens": map[string]any{
			"colors": map[string]any{
				"primary":    "#3366FF",
				"background": "#FFFFFF",
				"text":       "#111111",
			},
		},
		"constraints": map[string]any{
			"safeMargin": 0.05,
		},
		"layouts": []any{
			map[string]any{
				"name": "Title / Hero",
				"placeholders": []any{
					map[string]any{"id": "title", "type": "text", "geometry": map[string]any{"x": 0.1, "y": 0.2, "w": 0.8, "h": 0.2}},
					map[string]any{"id": "subtitle", "type": "text", "geometry": map[string]any{"x": 0.1, "y": 0.45, "w": 0.8, "h": 0.15}},
				},
			},
		},
	}
}
