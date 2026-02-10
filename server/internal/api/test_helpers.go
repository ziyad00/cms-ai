package api

import (
	"fmt"
	"net/http"

	"github.com/ziyad/cms-ai/server/internal/auth"
)

// addTestAuth adds a valid JWT token header to the request for testing
func addTestAuth(req *http.Request, userID, orgID string, role auth.Role) {
	token, err := auth.GenerateToken(userID, orgID, role)
	if err != nil {
		panic(fmt.Sprintf("failed to generate test token: %v", err))
	}
	req.Header.Set("Authorization", "Bearer "+token)
	// Log for debugging
	// log.Printf("[TEST DEBUG] Added Auth Header: %s", req.Header.Get("Authorization"))
}
