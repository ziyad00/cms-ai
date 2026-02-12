package api

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewID_returns_valid_uuid(t *testing.T) {
	id := newID("job")
	_, err := uuid.Parse(id)
	require.NoError(t, err, "newID must return a valid UUID, got %q", id)
}

func TestNewID_unique(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := newID("tpl")
		assert.False(t, ids[id], "newID generated duplicate: %s", id)
		ids[id] = true
	}
}

func TestNewID_prefix_ignored(t *testing.T) {
	// Prefix is accepted for backward compatibility but not used in the output
	id := newID("anything")
	_, err := uuid.Parse(id)
	require.NoError(t, err, "any prefix should still produce valid UUID, got %q", id)
}
