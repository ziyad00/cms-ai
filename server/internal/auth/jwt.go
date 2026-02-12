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
		log.Fatal("JWT_SECRET environment variable is required. Please set it before starting the server.")
	}
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters long for security")
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
	if authHeader == "" {
		return Identity{}, ErrUnauthenticated
	}

	// Extract token from "Bearer <token>"
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return Identity{}, ErrUnauthenticated
	}
	tokenString := authHeader[7:]

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return Identity{}, ErrUnauthenticated
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return Identity{
			UserID: claims.UserID,
			OrgID:  claims.OrgID,
			Role:   claims.Role,
		}, nil
	}

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
