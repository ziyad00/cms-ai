package spec

import "testing"

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
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got: %+v", errs)
	}
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
	if !found {
		t.Fatalf("expected overlap error, got: %+v", errs)
	}
}
