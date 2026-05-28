package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// ---------------------------------------------------------------------------
// Password helpers
// ---------------------------------------------------------------------------

// HashPassword hashes a plain-text password using bcrypt.
func HashPassword(plain string) (string, error) {
	if plain == "" {
		return "", errors.New("password cannot be empty")
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	return string(bytes), err
}

// CheckPassword compares a bcrypt hash against a plain-text candidate.
// Returns true when they match.
func CheckPassword(hash, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	return err == nil
}

// ---------------------------------------------------------------------------
// JWT helpers
// ---------------------------------------------------------------------------

// GenerateJWT mints a signed 24-hour access token.
// The JWT_SECRET env var is read at call time so hot-reloads work in dev.
func GenerateJWT(userId, role string, branchIds []string, email string, roles []string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET is not configured")
	}

	claims := JWTClaims{
		UserId:    userId,
		Role:      role,
		Roles:     roles,
		BranchIds: branchIds,
		Email:     email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "skin-firts-local",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseJWT validates and parses a token string, returning its claims.
func ParseJWT(tokenStr string) (*JWTClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET is not configured")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
