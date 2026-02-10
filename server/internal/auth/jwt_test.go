package auth

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	userID := "user-123"
	orgID := "org-456"
	role := RoleEditor

	token, err := GenerateToken(userID, orgID, role)

	require.NoError(t, err)
	assert.NotEmpty(t, token, "Token should not be empty")
	assert.Greater(t, len(token), 100, "Token should be reasonably long")

	// Token should have 3 parts separated by dots (header.payload.signature)
	parts := len(token)
	dotCount := 0
	for i := 0; i < parts; i++ {
		if token[i] == '.' {
			dotCount++
		}
	}
	assert.Equal(t, 2, dotCount, "JWT should have exactly 2 dots")
}

func TestJWTAuthenticator(t *testing.T) {
	authenticator := JWTAuthenticator{}

	// Generate a valid token for testing
	userID := "test-user-123"
	orgID := "test-org-456"
	role := RoleEditor

	validToken, err := GenerateToken(userID, orgID, role)
	require.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectError    bool
		expectedUserID string
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			expectError:    false,
			expectedUserID: userID,
		},
		{
			name:        "missing auth header",
			authHeader:  "",
			expectError: true,
		},
		{
			name:        "malformed auth header",
			authHeader:  "InvalidFormat",
			expectError: true,
		},
		{
			name:        "missing Bearer prefix",
			authHeader:  validToken,
			expectError: true,
		},
		{
			name:        "invalid token",
			authHeader:  "Bearer invalid-token",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			identity, err := authenticator.Authenticate(req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, ErrUnauthenticated, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUserID, identity.UserID)
				assert.Equal(t, orgID, identity.OrgID)
				assert.Equal(t, role, identity.Role)
			}
		})
	}
}

func TestHeaderAuthenticator(t *testing.T) {
	authenticator := HeaderAuthenticator{}

	tests := []struct {
		name        string
		userID      string
		orgID       string
		role        string
		expectError bool
	}{
		{
			name:        "valid headers",
			userID:      "test-user",
			orgID:       "test-org",
			role:        "Editor",
			expectError: false,
		},
		{
			name:        "missing user ID",
			userID:      "",
			orgID:       "test-org",
			role:        "Editor",
			expectError: true,
		},
		{
			name:        "missing org ID",
			userID:      "test-user",
			orgID:       "",
			role:        "Editor",
			expectError: true,
		},
		{
			name:        "missing role defaults to Editor",
			userID:      "test-user",
			orgID:       "test-org",
			role:        "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.userID != "" {
				req.Header.Set("X-User-Id", tt.userID)
			}
			if tt.orgID != "" {
				req.Header.Set("X-Org-Id", tt.orgID)
			}
			if tt.role != "" {
				req.Header.Set("X-Role", tt.role)
			}

			identity, err := authenticator.Authenticate(req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, ErrUnauthenticated, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.userID, identity.UserID)
				assert.Equal(t, tt.orgID, identity.OrgID)

				expectedRole := RoleEditor
				if tt.role != "" {
					expectedRole = Role(tt.role)
				}
				assert.Equal(t, expectedRole, identity.Role)
			}
		})
	}
}

func TestRoles(t *testing.T) {
	// Test role definitions exist
	assert.Equal(t, Role("Viewer"), RoleViewer)
	assert.Equal(t, Role("Editor"), RoleEditor)
	assert.Equal(t, Role("Admin"), RoleAdmin)
	assert.Equal(t, Role("Owner"), RoleOwner)
}

func TestIdentityContext(t *testing.T) {
	ctx := context.Background()
	identity := Identity{
		UserID: "user-123",
		OrgID:  "org-456",
		Role:   RoleEditor,
	}

	// Test setting identity in context
	ctxWithIdentity := WithIdentity(ctx, identity)
	assert.NotEqual(t, ctx, ctxWithIdentity, "Context should be different")

	// Test getting identity from context
	retrievedIdentity, ok := GetIdentity(ctxWithIdentity)
	assert.True(t, ok, "Should be able to get identity from context")
	assert.Equal(t, identity, retrievedIdentity, "Retrieved identity should match original")

	// Test getting identity from empty context
	_, ok = GetIdentity(ctx)
	assert.False(t, ok, "Should not find identity in empty context")
}