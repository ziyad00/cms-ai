package auth

import (
	"context"
	"errors"
	"net/http"
)

type Role string

const (
	RoleOwner  Role = "Owner"
	RoleAdmin  Role = "Admin"
	RoleEditor Role = "Editor"
	RoleViewer Role = "Viewer"
)

type Identity struct {
	UserID string
	OrgID  string
	Role   Role
}

type ctxKeyIdentity struct{}

func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, ctxKeyIdentity{}, id)
}

func GetIdentity(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(ctxKeyIdentity{}).(Identity)
	return id, ok
}

type Authenticator interface {
	Authenticate(r *http.Request) (Identity, error)
}

var ErrUnauthenticated = errors.New("unauthenticated")

type HeaderAuthenticator struct{}

// HeaderAuthenticator is a dev-friendly authenticator.
// Required headers: X-User-Id, X-Org-Id, optional X-Role.
func (HeaderAuthenticator) Authenticate(r *http.Request) (Identity, error) {
	userID := r.Header.Get("X-User-Id")
	orgID := r.Header.Get("X-Org-Id")
	if userID == "" || orgID == "" {
		return Identity{}, ErrUnauthenticated
	}
	role := Role(r.Header.Get("X-Role"))
	if role == "" {
		role = RoleEditor
	}
	return Identity{UserID: userID, OrgID: orgID, Role: role}, nil
}

func RequireRole(id Identity, min Role) bool {
	return roleRank(id.Role) >= roleRank(min)
}

func roleRank(r Role) int {
	switch r {
	case RoleOwner:
		return 4
	case RoleAdmin:
		return 3
	case RoleEditor:
		return 2
	case RoleViewer:
		return 1
	default:
		return 0
	}
}
