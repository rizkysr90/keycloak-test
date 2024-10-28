package utils

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims extends jwt.RegisteredClaims to include custom claims
type CustomClaims struct {
	// Add your custom claims here, for example:
	// Username string `json:"username,omitempty"`
	// Role     string `json:"role,omitempty"`
	jwt.RegisteredClaims
}
// TokenValidator handles JWT token validation
type TokenValidator struct {
	secretKey []byte
}
// NewTokenValidator creates a new TokenValidator instance
func NewTokenValidator(secretKey string) *TokenValidator {
	return &TokenValidator{
		secretKey: []byte(secretKey),
	}
}
// ValidateWithPublicKey validates a token using RSA public key
func ValidateWithPublicKey(tokenString string, publicKey *rsa.PublicKey) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}