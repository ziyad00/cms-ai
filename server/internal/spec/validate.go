package spec

import "fmt"

type Validator interface {
	Validate(spec TemplateSpec) []ValidationError
}

type DefaultValidator struct{}

func (v DefaultValidator) Validate(spec TemplateSpec) []ValidationError {
	var errors []ValidationError

	if spec.Tokens == nil {
		errors = append(errors, ValidationError{Path: "$.tokens", Message: "tokens is required"})
	}

	if len(spec.Layouts) == 0 {
		errors = append(errors, ValidationError{Path: "$.layouts", Message: "layouts must be a non-empty array"})
		return errors
	}

	safeMargin := spec.Constraints.SafeMargin
	if safeMargin == 0 {
		safeMargin = 0.05
	}
	if safeMargin < 0 || safeMargin >= 0.5 {
		errors = append(errors, ValidationError{Path: "$.constraints.safeMargin", Message: "safeMargin must be in [0, 0.5)"})
		safeMargin = 0.05
	}

	for layoutIndex, layout := range spec.Layouts {
		layoutPath := fmt.Sprintf("$.layouts[%d]", layoutIndex)

		if layout.Name == "" {
			errors = append(errors, ValidationError{Path: layoutPath + ".name", Message: "name is required"})
		}

		if len(layout.Placeholders) == 0 {
			errors = append(errors, ValidationError{Path: layoutPath + ".placeholders", Message: "placeholders must be non-empty"})
			continue
		}

		rects := make([]rect, 0, len(layout.Placeholders))
		for placeholderIndex, placeholder := range layout.Placeholders {
			placeholderPath := fmt.Sprintf("%s.placeholders[%d]", layoutPath, placeholderIndex)
			if placeholder.ID == "" {
				errors = append(errors, ValidationError{Path: placeholderPath + ".id", Message: "id is required"})
			}

			x, y, w, h := placeholder.Geometry.X, placeholder.Geometry.Y, placeholder.Geometry.W, placeholder.Geometry.H
			if w <= 0 || h <= 0 {
				errors = append(errors, ValidationError{Path: placeholderPath + ".geometry", Message: "w and h must be > 0"})
				continue
			}

			if x < safeMargin || y < safeMargin {
				errors = append(errors, ValidationError{Path: placeholderPath + ".geometry", Message: "x/y must respect safe margins"})
			}
			if x+w > 1.0-safeMargin || y+h > 1.0-safeMargin {
				errors = append(errors, ValidationError{Path: placeholderPath + ".geometry", Message: "geometry must fit within safe margins"})
			}

			rects = append(rects, rect{x: x, y: y, w: w, h: h, id: placeholder.ID})
		}

		for i := 0; i < len(rects); i++ {
			for j := i + 1; j < len(rects); j++ {
				if rectsOverlap(rects[i], rects[j]) {
					errors = append(errors, ValidationError{Path: layoutPath, Message: fmt.Sprintf("placeholders overlap: %s and %s", rects[i].id, rects[j].id)})
				}
			}
		}
	}

	return errors
}

type rect struct {
	x, y, w, h float64
	id         string
}

func rectsOverlap(a rect, b rect) bool {
	// Touching edges is allowed.
	if a.x+a.w <= b.x || b.x+b.w <= a.x {
		return false
	}
	if a.y+a.h <= b.y || b.y+b.h <= a.y {
		return false
	}
	return true
}
