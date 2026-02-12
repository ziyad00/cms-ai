package store

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONMap_Value_roundtrip(t *testing.T) {
	m := JSONMap{"filename": "export.pptx", "versionNo": "1"}

	val, err := m.Value()
	require.NoError(t, err)

	// Value() must return []byte (JSON), which is what pgx sends to Postgres
	b, ok := val.([]byte)
	require.True(t, ok, "Value() should return []byte, got %T", val)

	var scanned JSONMap
	err = scanned.Scan(b)
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

func TestJob_metadata_pointer_serialization(t *testing.T) {
	// This is the exact pattern used in production code (router_v1.go).
	// It must produce valid JSON bytes, not fail with pgx OID 0 error.
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

	b, ok := val.([]byte)
	require.True(t, ok, "must produce []byte for pgx, got %T", val)
	assert.Contains(t, string(b), "deck-export-v1.pptx")
}
