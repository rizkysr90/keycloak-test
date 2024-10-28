package utils

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWKKey `json:"keys"`
}

// JWKKey represents a JSON Web Key
type JWKKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"` // RSA public key modulus
	E   string `json:"e"` // RSA public key exponent
}


// KeyManager handles fetching and caching of JWKS
type KeyManager struct {
	jwksURL     string
	keyCache    map[string]*rsa.PublicKey
	cacheMutex  sync.RWMutex
	lastUpdate  time.Time
	cacheExpiry time.Duration
}

// NewKeyManager creates a new KeyManager instance
func NewKeyManager(jwksURL string) *KeyManager {
	return &KeyManager{
		jwksURL:     jwksURL,
		keyCache:    make(map[string]*rsa.PublicKey),
		cacheExpiry: 12 * time.Hour, // Cache keys for 12 hours
	}
}
// GetPublicKey fetches the public key for a given key ID
func (km *KeyManager) GetPublicKey(kid string) (*rsa.PublicKey, error) {
	// Check cache first
	km.cacheMutex.RLock()
	if key, exists := km.keyCache[kid]; exists && !km.isCacheExpired() {
		km.cacheMutex.RUnlock()
		return key, nil
	}
	km.cacheMutex.RUnlock()

	// Fetch new keys if cache miss or expired
	if err := km.refreshKeys(); err != nil {
		return nil, fmt.Errorf("failed to refresh keys: %w", err)
	}

	// Check cache again after refresh
	km.cacheMutex.RLock()
	defer km.cacheMutex.RUnlock()
	if key, exists := km.keyCache[kid]; exists {
		return key, nil
	}

	return nil, fmt.Errorf("key ID %s not found", kid)
}

// refreshKeys fetches fresh keys from the JWKS endpoint
func (km *KeyManager) refreshKeys() error {
	resp, err := http.Get(km.jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JWKS: status code %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	// Convert JWK to RSA public keys
	newCache := make(map[string]*rsa.PublicKey)
	for _, key := range jwks.Keys {
		if key.Kty != "RSA" || key.Use != "sig" || key.Alg != "RS256" {
			continue
		}

		publicKey, err := jwkToPublicKey(key)
		if err != nil {
			return fmt.Errorf("failed to convert JWK to public key: %w", err)
		}

		newCache[key.Kid] = publicKey
	}

	// Update cache
	km.cacheMutex.Lock()
	km.keyCache = newCache
	km.lastUpdate = time.Now()
	km.cacheMutex.Unlock()

	return nil
}
// jwkToPublicKey converts a JWK to an RSA public key
func jwkToPublicKey(jwk JWKKey) (*rsa.PublicKey, error) {
	// Decode the modulus
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode the exponent
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert bytes to big.Int for the modulus
	n := new(big.Int).SetBytes(nBytes)

	// Convert bytes to int for the exponent
	var e int
	for i := 0; i < len(eBytes); i++ {
		e = e<<8 + int(eBytes[i])
	}

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}

func (km *KeyManager) isCacheExpired() bool {
	return time.Since(km.lastUpdate) > km.cacheExpiry
}