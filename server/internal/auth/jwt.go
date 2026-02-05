package auth

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(getJWTSecret())

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Default secret for development - should be set in production
		return "dev-secret-change-in-production"
	}
	return secret
}

type JWTAuthenticator struct{}

type Claims struct {
	UserID string `json:"userId"`
	OrgID  string `json:"orgId"`
	Role   Role   `json:"role"`
	jwt.RegisteredClaims
}

// Authenticate validates JWT token from Authorization header
func (JWTAuthenticator) Authenticate(r *http.Request) (Identity, error) {
	authHeader := r.Header.Get("Authorization")
	log.Printf("[DEBUG] JWT Auth - Authorization header: %s", authHeader)
	if authHeader == "" {
		log.Printf("[DEBUG] JWT Auth - No authorization header")
		return Identity{}, ErrUnauthenticated
	}

	// Extract token from "Bearer <token>"
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		log.Printf("[DEBUG] JWT Auth - Invalid authorization header format: %s", authHeader)
		return Identity{}, ErrUnauthenticated
	}
	tokenString := authHeader[7:]
	tokenPreview := tokenString
	if len(tokenString) > 20 {
		tokenPreview = tokenString[:20] + "..."
	}
	log.Printf("[DEBUG] JWT Auth - Extracted token: %s", tokenPreview)

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("[DEBUG] JWT Auth - Invalid signing method: %v", token.Method)
			return nil, errors.New("invalid signing method")
		}
		log.Printf("[DEBUG] JWT Auth - Using HMAC signing method, secret length: %d", len(jwtSecret))
		return jwtSecret, nil
	})

	if err != nil {
		log.Printf("[DEBUG] JWT Auth - Token parsing failed: %v", err)
		return Identity{}, ErrUnauthenticated
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		log.Printf("[DEBUG] JWT Auth - Valid token for user: %s, org: %s", claims.UserID, claims.OrgID)
		return Identity{
			UserID: claims.UserID,
			OrgID:  claims.OrgID,
			Role:   claims.Role,
		}, nil
	}

	log.Printf("[DEBUG] JWT Auth - Token claims invalid or token not valid")
	return Identity{}, ErrUnauthenticated
}

// GenerateToken creates a JWT token for a user
func GenerateToken(userID, orgID string, role Role) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour * 7) // 7 days

	claims := &Claims{
		UserID: userID,
		OrgID:  orgID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
