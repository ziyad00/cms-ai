package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziyad/cms-ai/server/internal/auth"
	"github.com/ziyad/cms-ai/server/internal/store"
)

func TestDeckExportList_IncludesQueuedJobs(t *testing.T) {
	s := NewServer()
	h := s.Handler()
	ctx := context.Background()

	// 1. Create a deck
	deck := store.Deck{
		ID:    "test-deck-1",
		OrgID: "org-1",
		Name:  "Test Deck",
	}
	_, err := s.Store.Decks().CreateDeck(ctx, deck)
	require.NoError(t, err)

	// 2. Create a deck version
	version := store.DeckVersion{
		ID:        "version-1",
		Deck:      "test-deck-1",
		OrgID:     "org-1",
		VersionNo: 1,
		SpecJSON:  []byte(`{"slides": []}`),
	}
	_, err = s.Store.Decks().CreateDeckVersion(ctx, version)
	require.NoError(t, err)

	// 3. Manually enqueue a QUEUED job for this version
	job := store.Job{
		ID:       "job-queued-1",
		OrgID:    "org-1",
		Type:     store.JobExport,
		Status:   store.JobQueued,
		InputRef: "version-1",
	}
	_, err = s.Store.Jobs().Enqueue(ctx, job)
	require.NoError(t, err)

	// 4. Call handleListDeckExports
	req := httptest.NewRequest(http.MethodGet, "/v1/decks/test-deck-1/exports", nil)
	addTestAuth(req, "user-1", "org-1", auth.RoleEditor)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	// 5. Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var resp struct {
		Exports []store.Job `json:"exports"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Should include the queued job
	assert.Len(t, resp.Exports, 1, "Should return 1 export job")
	assert.Equal(t, "job-queued-1", resp.Exports[0].ID)
	assert.Equal(t, store.JobQueued, resp.Exports[0].Status)
}

func TestDeckExportList_AggregatesMultipleVersions(t *testing.T) {
	s := NewServer()
	h := s.Handler()
	ctx := context.Background()

	// 1. Create a deck
	deckID := "multi-ver-deck"
	deck := store.Deck{ID: deckID, OrgID: "org-1", Name: "Multi-version Deck"}
	_, _ = s.Store.Decks().CreateDeck(ctx, deck)

	// 2. Create two versions
	v1 := store.DeckVersion{ID: "v1", Deck: deckID, OrgID: "org-1", VersionNo: 1, SpecJSON: []byte(`{}`)}
	v2 := store.DeckVersion{ID: "v2", Deck: deckID, OrgID: "org-1", VersionNo: 2, SpecJSON: []byte(`{}`)}
	_, _ = s.Store.Decks().CreateDeckVersion(ctx, v1)
	_, _ = s.Store.Decks().CreateDeckVersion(ctx, v2)

	// 3. Create jobs for each version
	j1 := store.Job{ID: "job-v1", OrgID: "org-1", Type: store.JobExport, Status: store.JobDone, InputRef: "v1"}
	j2 := store.Job{ID: "job-v2", OrgID: "org-1", Type: store.JobExport, Status: store.JobQueued, InputRef: "v2"}
	_, _ = s.Store.Jobs().Enqueue(ctx, j1)
	_, _ = s.Store.Jobs().Enqueue(ctx, j2)

	// 4. List exports
	req := httptest.NewRequest(http.MethodGet, "/v1/decks/"+deckID+"/exports", nil)
	addTestAuth(req, "user-1", "org-1", auth.RoleEditor)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	// 5. Verify
	var resp struct {
		Exports []store.Job `json:"exports"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp.Exports, 2, "Should return jobs from both versions")
}
