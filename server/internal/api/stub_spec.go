package api

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
