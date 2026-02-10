package auth

import (
	"net/http"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name     string
		userRole Role
		minRole  Role
		want     bool
	}{
		{"Owner vs Owner", RoleOwner, RoleOwner, true},
		{"Owner vs Admin", RoleOwner, RoleAdmin, true},
		{"Owner vs Viewer", RoleOwner, RoleViewer, true},
		
		{"Admin vs Owner", RoleAdmin, RoleOwner, false},
		{"Admin vs Admin", RoleAdmin, RoleAdmin, true},
		{"Admin vs Editor", RoleAdmin, RoleEditor, true},
		
		{"Editor vs Admin", RoleEditor, RoleAdmin, false},
		{"Editor vs Editor", RoleEditor, RoleEditor, true},
		{"Editor vs Viewer", RoleEditor, RoleViewer, true},
		
		{"Viewer vs Editor", RoleViewer, RoleEditor, false},
		{"Viewer vs Viewer", RoleViewer, RoleViewer, true},
		
		{"Unknown role", Role("None"), RoleViewer, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := Identity{Role: tt.userRole}
			assert.Equal(t, tt.want, RequireRole(id, tt.minRole))
		})
	}
}

func TestRoleRank(t *testing.T) {
	assert.Equal(t, 4, roleRank(RoleOwner))
	assert.Equal(t, 3, roleRank(RoleAdmin))
	assert.Equal(t, 2, roleRank(RoleEditor))
	assert.Equal(t, 1, roleRank(RoleViewer))
	assert.Equal(t, 0, roleRank(Role("Other")))
}

func TestJWTAuthFlow(t *testing.T) {
	userID := "test-user"
	orgID := "test-org"
	role := RoleEditor

	// 1. Generate token
	token, err := GenerateToken(userID, orgID, role)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// 2. Authenticate
	auth := JWTAuthenticator{}
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	id, err := auth.Authenticate(req)
	require.NoError(t, err, "Authenticator should accept valid token")
	assert.Equal(t, userID, id.UserID)
	assert.Equal(t, orgID, id.OrgID)
	assert.Equal(t, role, id.Role)
}
