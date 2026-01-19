package api

import (
	"net/http"

	"github.com/ziyad/cms-ai/server/internal/auth"
)

type ctxKeyIdentity struct{}

func withAuth(a auth.Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := a.Authenticate(r)
			if err != nil {
				writeError(w, r, http.StatusUnauthorized, "unauthorized")
				return
			}
			r = r.WithContext(auth.WithIdentity(r.Context(), id))
			next.ServeHTTP(w, r)
		})
	}
}
