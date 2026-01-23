package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziyad/cms-ai/server/internal/spec"
	"github.com/ziyad/cms-ai/server/internal/store"
)

type mockStore struct {
	templates map[string]store.Template
	versions  map[string]store.TemplateVersion
	brandKits map[string]store.BrandKit
	metering  []store.MeteringEvent
}

func newMockStore() *mockStore {
	return &mockStore{
		templates: make(map[string]store.Template),
		versions:  make(map[string]store.TemplateVersion),
		brandKits: make(map[string]store.BrandKit),
		metering:  make([]store.MeteringEvent, 0),
	}
}

func (m *mockStore) Templates() store.TemplateStore {
	return &mockTemplateStore{templates: m.templates, versions: m.versions}
}

func (m *mockStore) BrandKits() store.BrandKitStore {
	return &mockBrandKitStore{brandKits: m.brandKits}
}

func (m *mockStore) Metering() store.MeteringStore {
	return &mockMeteringStore{metering: &m.metering}
}

func (m *mockStore) Decks() store.DeckStore                 { return nil }
func (m *mockStore) Assets() store.AssetStore               { return nil }
func (m *mockStore) Jobs() store.JobStore                   { return nil }
func (m *mockStore) Audit() store.AuditStore                { return nil }
func (m *mockStore) Users() store.UserStore                 { return nil }
func (m *mockStore) Organizations() store.OrganizationStore { return nil }

type mockTemplateStore struct {
	templates map[string]store.Template
	versions  map[string]store.TemplateVersion
}

func (m *mockTemplateStore) CreateTemplate(ctx context.Context, t store.Template) (store.Template, error) {
	m.templates[t.ID] = t
	return t, nil
}

func (m *mockTemplateStore) ListTemplates(ctx context.Context, orgID string) ([]store.Template, error) {
	var templates []store.Template
	for _, t := range m.templates {
		if t.OrgID == orgID {
			templates = append(templates, t)
		}
	}
	return templates, nil
}

func (m *mockTemplateStore) GetTemplate(ctx context.Context, orgID, id string) (store.Template, bool, error) {
	t, exists := m.templates[id]
	if !exists || t.OrgID != orgID {
		return store.Template{}, false, nil
	}
	return t, true, nil
}

func (m *mockTemplateStore) UpdateTemplate(ctx context.Context, t store.Template) (store.Template, error) {
	m.templates[t.ID] = t
	return t, nil
}

func (m *mockTemplateStore) CreateVersion(ctx context.Context, v store.TemplateVersion) (store.TemplateVersion, error) {
	m.versions[v.ID] = v
	return v, nil
}

func (m *mockTemplateStore) ListVersions(ctx context.Context, orgID, templateID string) ([]store.TemplateVersion, error) {
	var versions []store.TemplateVersion
	for _, v := range m.versions {
		if v.OrgID == orgID && v.Template == templateID {
			versions = append(versions, v)
		}
	}
	return versions, nil
}

func (m *mockTemplateStore) GetVersion(ctx context.Context, orgID, versionID string) (store.TemplateVersion, bool, error) {
	v, exists := m.versions[versionID]
	if !exists || v.OrgID != orgID {
		return store.TemplateVersion{}, false, nil
	}
	return v, true, nil
}

type mockBrandKitStore struct {
	brandKits map[string]store.BrandKit
}

func (m *mockBrandKitStore) Create(ctx context.Context, b store.BrandKit) (store.BrandKit, error) {
	m.brandKits[b.ID] = b
	return b, nil
}

func (m *mockBrandKitStore) List(ctx context.Context, orgID string) ([]store.BrandKit, error) {
	var brandKits []store.BrandKit
	for _, bk := range m.brandKits {
		if bk.OrgID == orgID {
			brandKits = append(brandKits, bk)
		}
	}
	return brandKits, nil
}

type mockMeteringStore struct {
	metering *[]store.MeteringEvent
}

func (m *mockMeteringStore) Record(ctx context.Context, e store.MeteringEvent) (store.MeteringEvent, error) {
	*m.metering = append(*m.metering, e)
	return e, nil
}

func (m *mockMeteringStore) SumByType(ctx context.Context, orgID string, eventType string) (int, error) {
	sum := 0
	for _, e := range *m.metering {
		if e.OrgID == orgID && e.Type == eventType {
			sum += e.Quantity
		}
	}
	return sum, nil
}

// Mock orchestrator for testing
type mockOrchestrator struct {
	response *GenerationResponse
	err      error
}

func (m *mockOrchestrator) GenerateTemplateSpec(ctx context.Context, req GenerationRequest) (*GenerationResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func (m *mockOrchestrator) RepairTemplateSpec(ctx context.Context, invalidSpec *spec.TemplateSpec, errors []spec.ValidationError) (*spec.TemplateSpec, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response.Spec, nil
}

func TestAIService_NewAIService(t *testing.T) {
	store := newMockStore()
	service := NewAIService(store)

	assert.NotNil(t, service)
	assert.NotNil(t, service.orchestrator)
	assert.NotNil(t, service.store)
}

func TestAIService_GenerateTemplateForRequest(t *testing.T) {
	mockStore := newMockStore()

	// Add a brand kit
	brandKit := store.BrandKit{
		ID:    "bk-1",
		OrgID: "org-1",
		Name:  "Test Brand",
		Tokens: map[string]any{
			"colors": map[string]any{
				"primary": "#FF0000",
			},
		},
	}
	_, _ = mockStore.BrandKits().Create(context.Background(), brandKit)

	// Create service with mock orchestrator
	expectedSpec := &spec.TemplateSpec{
		Tokens: map[string]any{
			"colors": map[string]any{
				"primary": "#FF0000",
			},
		},
		Constraints: spec.Constraints{SafeMargin: 0.05},
		Layouts: []spec.Layout{
			{
				Name: "Title Slide",
				Placeholders: []spec.Placeholder{
					{
						ID:       "title",
						Type:     "text",
						Geometry: spec.Geometry{X: 0.1, Y: 0.1, W: 0.8, H: 0.2},
					},
				},
			},
		},
	}

	mockOrch := &mockOrchestrator{
		response: &GenerationResponse{
			Spec:       expectedSpec,
			TokenUsage: 100,
			Cost:       0.001,
			Model:      "test-model",
		},
	}

	service := &AIService{
		orchestrator: mockOrch,
		store:        mockStore,
	}

	req := GenerationRequest{
		Prompt:     "Create a test presentation",
		Language:   "English",
		Tone:       "Professional",
		RTL:        false,
		BrandKitID: "bk-1",
	}

	spec, resp, err := service.GenerateTemplateForRequest(context.Background(), "org-1", "user-1", req, "bk-1")

	require.NoError(t, err)
	assert.NotNil(t, spec)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedSpec, spec)
	assert.Equal(t, 100, resp.TokenUsage)
	assert.Equal(t, 0.001, resp.Cost)
	assert.Equal(t, "test-model", resp.Model)

	// Check that metering was recorded
	metering := mockStore.Metering().(*mockMeteringStore)
	assert.Len(t, *metering.metering, 1)
	assert.Equal(t, "org-1", (*metering.metering)[0].OrgID)
	assert.Equal(t, "user-1", (*metering.metering)[0].UserID)
	assert.Equal(t, "ai_generation", (*metering.metering)[0].Type)
	assert.Equal(t, 100, (*metering.metering)[0].Quantity)
}

func TestAIService_GenerateTemplateForRequest_NoBrandKit(t *testing.T) {
	mockStore := newMockStore()

	expectedSpec := &spec.TemplateSpec{
		Tokens:      map[string]any{},
		Constraints: spec.Constraints{SafeMargin: 0.05},
		Layouts:     []spec.Layout{},
	}

	mockOrch := &mockOrchestrator{
		response: &GenerationResponse{
			Spec:       expectedSpec,
			TokenUsage: 50,
			Cost:       0.0005,
			Model:      "test-model",
		},
	}

	service := &AIService{
		orchestrator: mockOrch,
		store:        mockStore,
	}

	req := GenerationRequest{
		Prompt:   "Create a test presentation",
		Language: "English",
		Tone:     "Professional",
		RTL:      false,
	}

	spec, resp, err := service.GenerateTemplateForRequest(context.Background(), "org-1", "user-1", req, "")

	require.NoError(t, err)
	assert.NotNil(t, spec)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedSpec, spec)
	assert.Equal(t, 50, resp.TokenUsage)
	assert.Equal(t, 0.0005, resp.Cost)
}

func TestAIService_GenerateTemplateForRequest_Error(t *testing.T) {
	mockStore := newMockStore()

	mockOrch := &mockOrchestrator{
		err: assert.AnError,
	}

	service := &AIService{
		orchestrator: mockOrch,
		store:        mockStore,
	}

	req := GenerationRequest{
		Prompt: "Create a test presentation",
		RTL:    false,
	}

	spec, resp, err := service.GenerateTemplateForRequest(context.Background(), "org-1", "user-1", req, "")

	assert.Error(t, err)
	assert.Nil(t, spec)
	assert.Nil(t, resp)
}
