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
