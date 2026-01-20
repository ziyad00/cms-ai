package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultValidator_ValidSpec(t *testing.T) {
	v := DefaultValidator{}

	s := TemplateSpec{
		Tokens:      map[string]any{"colors": map[string]any{"primary": "#3366FF"}},
		Constraints: Constraints{SafeMargin: 0.05},
		Layouts: []Layout{{
			Name: "Title",
			Placeholders: []Placeholder{
				{ID: "title", Geometry: Geometry{X: 0.1, Y: 0.2, W: 0.8, H: 0.2}},
				{ID: "subtitle", Geometry: Geometry{X: 0.1, Y: 0.45, W: 0.8, H: 0.15}},
			},
		}},
	}

	errs := v.Validate(s)
	assert.Len(t, errs, 0, "expected no errors for valid spec")
}

func TestDefaultValidator_Overlap(t *testing.T) {
	v := DefaultValidator{}

	s := TemplateSpec{
		Tokens:      map[string]any{"colors": map[string]any{"primary": "#3366FF"}},
		Constraints: Constraints{SafeMargin: 0.05},
		Layouts: []Layout{{
			Name: "Bad",
			Placeholders: []Placeholder{
				{ID: "a", Geometry: Geometry{X: 0.1, Y: 0.2, W: 0.6, H: 0.3}},
				{ID: "b", Geometry: Geometry{X: 0.5, Y: 0.3, W: 0.4, H: 0.3}},
			},
		}},
	}

	errs := v.Validate(s)
	found := false
	for _, e := range errs {
		if e.Path == "$.layouts[0]" && e.Message != "" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected overlap error, got: %+v", errs)
}

func TestDefaultValidator_MissingTokens(t *testing.T) {
	v := DefaultValidator{}

	s := TemplateSpec{
		Constraints: Constraints{SafeMargin: 0.05},
		Layouts: []Layout{{
			Name: "Title",
			Placeholders: []Placeholder{
				{ID: "title", Geometry: Geometry{X: 0.1, Y: 0.2, W: 0.8, H: 0.2}},
			},
		}},
	}

	errs := v.Validate(s)
	assert.Len(t, errs, 1)
	assert.Equal(t, "$.tokens", errs[0].Path)
	assert.Contains(t, errs[0].Message, "tokens is required")
}

func TestDefaultValidator_EmptyLayouts(t *testing.T) {
	v := DefaultValidator{}

	s := TemplateSpec{
		Tokens:      map[string]any{"colors": map[string]any{"primary": "#3366FF"}},
		Constraints: Constraints{SafeMargin: 0.05},
		Layouts:     []Layout{},
	}

	errs := v.Validate(s)
	require.Len(t, errs, 1)
	assert.Equal(t, "$.layouts", errs[0].Path)
	assert.Contains(t, errs[0].Message, "layouts must be a non-empty array")
}

func TestDefaultValidator_SafeMarginValidation(t *testing.T) {
	v := DefaultValidator{}

	tests := []struct {
		name       string
		safeMargin float64
		expectErr  bool
	}{
		{"valid margin", 0.05, false},
		{"zero margin", 0.0, false},
		{"negative margin", -0.1, true},
		{"too large margin", 0.5, true},
		{"edge case valid", 0.49, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TemplateSpec{
				Tokens:      map[string]any{"colors": map[string]any{"primary": "#3366FF"}},
				Constraints: Constraints{SafeMargin: tt.safeMargin},
				Layouts: []Layout{{
					Name: "Title",
					Placeholders: []Placeholder{
						{ID: "title", Geometry: Geometry{X: 0.1, Y: 0.2, W: 0.8, H: 0.2}},
					},
				}},
			}

			errs := v.Validate(s)
			hasMarginError := false
			for _, err := range errs {
				if err.Path == "$.constraints.safeMargin" {
					hasMarginError = true
					break
				}
			}

			if tt.expectErr {
				assert.True(t, hasMarginError, "expected safe margin error")
			} else {
				assert.False(t, hasMarginError, "unexpected safe margin error: %+v", errs)
			}
		})
	}
}

func TestDefaultValidator_MissingLayoutName(t *testing.T) {
	v := DefaultValidator{}

	s := TemplateSpec{
		Tokens:      map[string]any{"colors": map[string]any{"primary": "#3366FF"}},
		Constraints: Constraints{SafeMargin: 0.05},
		Layouts: []Layout{{
			Name: "",
			Placeholders: []Placeholder{
				{ID: "title", Geometry: Geometry{X: 0.1, Y: 0.2, W: 0.8, H: 0.2}},
			},
		}},
	}

	errs := v.Validate(s)
	found := false
	for _, e := range errs {
		if e.Path == "$.layouts[0].name" && e.Message == "name is required" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected missing layout name error")
}

func TestDefaultValidator_EmptyPlaceholders(t *testing.T) {
	v := DefaultValidator{}

	s := TemplateSpec{
		Tokens:      map[string]any{"colors": map[string]any{"primary": "#3366FF"}},
		Constraints: Constraints{SafeMargin: 0.05},
		Layouts: []Layout{{
			Name:         "Title",
			Placeholders: []Placeholder{},
		}},
	}

	errs := v.Validate(s)
	found := false
	for _, e := range errs {
		if e.Path == "$.layouts[0].placeholders" && e.Message == "placeholders must be non-empty" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected empty placeholders error")
}

func TestDefaultValidator_MissingPlaceholderID(t *testing.T) {
	v := DefaultValidator{}

	s := TemplateSpec{
		Tokens:      map[string]any{"colors": map[string]any{"primary": "#3366FF"}},
		Constraints: Constraints{SafeMargin: 0.05},
		Layouts: []Layout{{
			Name: "Title",
			Placeholders: []Placeholder{
				{ID: "", Geometry: Geometry{X: 0.1, Y: 0.2, W: 0.8, H: 0.2}},
			},
		}},
	}

	errs := v.Validate(s)
	found := false
	for _, e := range errs {
		if e.Path == "$.layouts[0].placeholders[0].id" && e.Message == "id is required" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected missing placeholder ID error")
}

func TestDefaultValidator_InvalidGeometrySize(t *testing.T) {
	v := DefaultValidator{}

	tests := []struct {
		name     string
		geometry Geometry
		expectErr bool
	}{
		{"valid size", Geometry{X: 0.1, Y: 0.2, W: 0.8, H: 0.2}, false},
		{"zero width", Geometry{X: 0.1, Y: 0.2, W: 0.0, H: 0.2}, true},
		{"zero height", Geometry{X: 0.1, Y: 0.2, W: 0.8, H: 0.0}, true},
		{"negative width", Geometry{X: 0.1, Y: 0.2, W: -0.1, H: 0.2}, true},
		{"negative height", Geometry{X: 0.1, Y: 0.2, W: 0.8, H: -0.1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TemplateSpec{
				Tokens:      map[string]any{"colors": map[string]any{"primary": "#3366FF"}},
				Constraints: Constraints{SafeMargin: 0.05},
				Layouts: []Layout{{
					Name: "Title",
					Placeholders: []Placeholder{
						{ID: "test", Geometry: tt.geometry},
					},
				}},
			}

			errs := v.Validate(s)
			hasSizeError := false
			for _, err := range errs {
				if err.Path == "$.layouts[0].placeholders[0].geometry" && err.Message == "w and h must be > 0" {
					hasSizeError = true
					break
				}
			}

			if tt.expectErr {
				assert.True(t, hasSizeError, "expected geometry size error")
			} else {
				assert.False(t, hasSizeError, "unexpected geometry size error: %+v", errs)
			}
		})
	}
}

func TestDefaultValidator_GeometryOutOfBounds(t *testing.T) {
	v := DefaultValidator{}

	tests := []struct {
		name       string
		geometry   Geometry
		safeMargin float64
		expectErr  bool
	}{
		{"valid within bounds", Geometry{X: 0.1, Y: 0.1, W: 0.8, H: 0.8}, 0.05, false},
		{"x too close to left", Geometry{X: 0.01, Y: 0.1, W: 0.8, H: 0.8}, 0.05, true},
		{"y too close to top", Geometry{X: 0.1, Y: 0.01, W: 0.8, H: 0.8}, 0.05, true},
		{"extends beyond right", Geometry{X: 0.1, Y: 0.1, W: 0.9, H: 0.8}, 0.05, true},
		{"extends beyond bottom", Geometry{X: 0.1, Y: 0.1, W: 0.8, H: 0.9}, 0.05, true},
		{"edge case valid", Geometry{X: 0.05, Y: 0.05, W: 0.85, H: 0.85}, 0.05, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TemplateSpec{
				Tokens:      map[string]any{"colors": map[string]any{"primary": "#3366FF"}},
				Constraints: Constraints{SafeMargin: tt.safeMargin},
				Layouts: []Layout{{
					Name: "Title",
					Placeholders: []Placeholder{
						{ID: "test", Geometry: tt.geometry},
					},
				}},
			}

			errs := v.Validate(s)
			hasBoundsError := false
			for _, err := range errs {
				if err.Path == "$.layouts[0].placeholders[0].geometry" &&
				   (err.Message == "x/y must respect safe margins" || err.Message == "geometry must fit within safe margins") {
					hasBoundsError = true
					break
				}
			}

			if tt.expectErr {
				assert.True(t, hasBoundsError, "expected geometry bounds error")
			} else {
				assert.False(t, hasBoundsError, "unexpected geometry bounds error: %+v", errs)
			}
		})
	}
}

func TestRectsOverlap(t *testing.T) {
	tests := []struct {
		name     string
		a        rect
		b        rect
		expected bool
	}{
		{
			name:     "no overlap - separated horizontally",
			a:        rect{x: 0.1, y: 0.1, w: 0.3, h: 0.3, id: "a"},
			b:        rect{x: 0.5, y: 0.1, w: 0.3, h: 0.3, id: "b"},
			expected: false,
		},
		{
			name:     "no overlap - separated vertically",
			a:        rect{x: 0.1, y: 0.1, w: 0.3, h: 0.3, id: "a"},
			b:        rect{x: 0.1, y: 0.5, w: 0.3, h: 0.3, id: "b"},
			expected: false,
		},
		{
			name:     "touching edges - no overlap",
			a:        rect{x: 0.1, y: 0.1, w: 0.3, h: 0.3, id: "a"},
			b:        rect{x: 0.4, y: 0.1, w: 0.3, h: 0.3, id: "b"},
			expected: false,
		},
		{
			name:     "clear overlap",
			a:        rect{x: 0.1, y: 0.1, w: 0.5, h: 0.5, id: "a"},
			b:        rect{x: 0.3, y: 0.3, w: 0.5, h: 0.5, id: "b"},
			expected: true,
		},
		{
			name:     "one contains the other",
			a:        rect{x: 0.1, y: 0.1, w: 0.8, h: 0.8, id: "a"},
			b:        rect{x: 0.2, y: 0.2, w: 0.3, h: 0.3, id: "b"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rectsOverlap(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)

			// Test symmetry
			result2 := rectsOverlap(tt.b, tt.a)
			assert.Equal(t, tt.expected, result2, "overlap should be symmetric")
		})
	}
}
