package assets

import _ "embed"

//go:embed python_renderer.py
var embeddedPythonScript string

// GetEmbeddedPythonScript returns the embedded Python script content
func GetEmbeddedPythonScript() string {
	return embeddedPythonScript
}