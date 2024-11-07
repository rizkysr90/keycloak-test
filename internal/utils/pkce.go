package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// GenerateCodeVerifier creates a random code verifier for PKCE
func GenerateCodeVerifier() (string, error) {
    // Generate random bytes (32 bytes = 256 bits)
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    // Convert to URL-safe base64 string
    return base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateCodeChallenge creates SHA256 code challenge from verifier
func GenerateCodeChallenge(verifier string) string {
    // Create SHA256 hash of verifier
    hash := sha256.Sum256([]byte(verifier))
    // Convert to URL-safe base64 string
    return base64.RawURLEncoding.EncodeToString(hash[:])
}