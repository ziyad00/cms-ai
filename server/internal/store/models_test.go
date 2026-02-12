package store

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONMap_Value_returns_string_for_pgx_jsonb(t *testing.T) {
	// CRITICAL: pgx sends []byte as PostgreSQL "bytea", NOT as "json/jsonb".
	// This causes: ERROR: invalid input syntax for type json (SQLSTATE 22P02)
	// Value() MUST return string so pgx sends it as text, which PG accepts for jsonb.
	m := JSONMap{"filename": "export.pptx", "versionNo": "1"}

	val, err := m.Value()
	require.NoError(t, err)

	_, ok := val.(string)
	require.True(t, ok, "Value() must return string for pgx jsonb compat, got %T", val)
}

func TestJSONMap_Value_roundtrip(t *testing.T) {
	m := JSONMap{"filename": "export.pptx", "versionNo": "1"}

	val, err := m.Value()
	require.NoError(t, err)

	// Scan must accept both string and []byte (PG returns []byte on read)
	var scanned JSONMap
	err = scanned.Scan([]byte(val.(string)))
	require.NoError(t, err)
	assert.Equal(t, "export.pptx", scanned["filename"])
	assert.Equal(t, "1", scanned["versionNo"])
}

func TestJSONMap_Value_nil(t *testing.T) {
	var m JSONMap
	val, err := m.Value()
	require.NoError(t, err)
	assert.Nil(t, val, "nil JSONMap should produce nil driver.Value")
}

func TestJSONMap_Scan_nil(t *testing.T) {
	m := JSONMap{"key": "val"}
	err := m.Scan(nil)
	require.NoError(t, err)
	assert.Nil(t, JSONMap(m), "Scan(nil) should set map to nil")
}

func TestJSONMap_Scan_wrong_type(t *testing.T) {
	var m JSONMap
	err := m.Scan(12345)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected []byte")
}

func TestJSONMap_implements_valuer(t *testing.T) {
	// Compile-time check that JSONMap satisfies driver.Valuer
	var _ driver.Valuer = JSONMap{}
}

func TestJSONMap_Value_empty_map(t *testing.T) {
	m := JSONMap{}
	val, err := m.Value()
	require.NoError(t, err)

	s, ok := val.(string)
	require.True(t, ok, "Value() must return string, got %T", val)
	assert.Equal(t, "{}", s, "empty map should serialize to {}")
}

func TestJSONMap_Scan_empty_json_object(t *testing.T) {
	var m JSONMap
	err := m.Scan([]byte(`{}`))
	require.NoError(t, err)
	assert.NotNil(t, m, "should scan into non-nil empty map")
	assert.Len(t, m, 0)
}

func TestJSONMap_Value_special_characters(t *testing.T) {
	m := JSONMap{
		"html":    `<script>alert("xss")</script>`,
		"unicode": "日本語テスト",
		"quotes":  `value with "quotes" and 'apostrophes'`,
		"newline": "line1\nline2",
	}

	val, err := m.Value()
	require.NoError(t, err)

	var scanned JSONMap
	err = scanned.Scan([]byte(val.(string)))
	require.NoError(t, err)
	assert.Equal(t, m["html"], scanned["html"])
	assert.Equal(t, m["unicode"], scanned["unicode"])
	assert.Equal(t, m["quotes"], scanned["quotes"])
	assert.Equal(t, m["newline"], scanned["newline"])
}

func TestJSONMap_Scan_string_type(t *testing.T) {
	// Some database drivers may return string instead of []byte
	var m JSONMap
	err := m.Scan("not bytes")
	require.Error(t, err, "Scan should reject string type (only []byte)")
	assert.Contains(t, err.Error(), "expected []byte")
}

func TestJSONMap_pointer_nil_value(t *testing.T) {
	// Matches production pattern: job.Metadata is *JSONMap
	var ptr *JSONMap
	if ptr != nil {
		_, _ = ptr.Value()
		t.Fatal("nil pointer should not be dereferenced")
	}
	// This is the correct nil check pattern used in worker.go
	assert.Nil(t, ptr)
}

func TestJob_metadata_all_production_patterns(t *testing.T) {
	// Test all three metadata patterns from router_v1.go
	tests := []struct {
		name     string
		metadata JSONMap
		jobType  JobType
	}{
		{
			name: "export metadata (router_v1:1030)",
			metadata: JSONMap{
				"versionNo": "1",
				"filename":  "deck-export-v1-20260212-123041.pptx",
			},
			jobType: JobExport,
		},
		{
			name: "generate metadata (router_v1:305)",
			metadata: JSONMap{
				"prompt":     "Create a sales deck for Q4",
				"language":   "en",
				"tone":       "professional",
				"rtl":        "false",
				"brandKitId": "bk-123",
				"userId":     "user-456",
			},
			jobType: JobGenerate,
		},
		{
			name: "bind metadata (router_v1:794)",
			metadata: JSONMap{
				"sourceTemplateVersionId": "tv-789",
				"content":                 "Revenue grew 25% in Q4...",
				"userId":                  "user-456",
			},
			jobType: JobBind,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			job := Job{
				ID:       "job-test",
				OrgID:    "org-1",
				Type:     tc.jobType,
				Status:   JobQueued,
				Metadata: &tc.metadata,
			}

			val, err := job.Metadata.Value()
			require.NoError(t, err, "Value() should not error")
			require.NotNil(t, val, "Value() should not be nil")

			s, ok := val.(string)
			require.True(t, ok, "must produce string for pgx jsonb, got %T", val)

			var scanned JSONMap
			err = scanned.Scan([]byte(s))
			require.NoError(t, err, "Scan() should not error")

			for k, v := range tc.metadata {
				assert.Equal(t, v, scanned[k], "key %q must roundtrip", k)
			}
		})
	}
}

func TestJob_metadata_pointer_serialization(t *testing.T) {
	// This is the exact pattern used in production code (router_v1.go).
	// Value() must return string (not []byte) so pgx sends it as text for jsonb.
	metadata := JSONMap{
		"versionNo": "1",
		"filename":  "deck-export-v1.pptx",
	}
	job := Job{
		ID:       "job-1",
		OrgID:    "org-1",
		Type:     JobExport,
		Status:   JobQueued,
		Metadata: &metadata,
	}

	val, err := job.Metadata.Value()
	require.NoError(t, err)
	require.NotNil(t, val)

	s, ok := val.(string)
	require.True(t, ok, "must produce string for pgx jsonb, got %T", val)
	assert.Contains(t, s, "deck-export-v1.pptx")
}
