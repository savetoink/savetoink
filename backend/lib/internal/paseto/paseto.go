// Package paseto provides PASETO v4.local token generation and validation
// for the savetoink application authentication.
package paseto

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
)

const (
	// TokenPrefix is the prefix of v4.local PASETO tokens.
	TokenPrefix = "v4.local."

	// keySize is the required size in bytes for a v4 symmetric key.
	keySize = 32

	// DefaultAccessTTL is the default time-to-live for access tokens.
	DefaultAccessTTL = 1 * time.Hour

	// DefaultRefreshTTL is the default time-to-live for refresh tokens.
	DefaultRefreshTTL = 7 * 24 * time.Hour
)

// TokenClaims represents the claims stored in PASETO tokens.
type TokenClaims struct {
	// Subject is the account ID extracted from Auth0.
	Subject string

	// Email is the user's email address.
	Email string

	// KeyVersion identifies which key was used to encrypt the token.
	KeyVersion string

	// TokenType is either "access" or "refresh".
	TokenType string
}

// TokenPair holds an access and refresh token pair.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// KeyStore manages PASETO symmetric keys with support for multiple versions.
type KeyStore struct {
	currentKey     paseto.V4SymmetricKey
	currentVersion string
	previousKeys   map[string]paseto.V4SymmetricKey
	accessTTL      time.Duration
	refreshTTL     time.Duration
}

// KeyStoreConfig holds configuration for creating a new KeyStore.
type KeyStoreConfig struct {
	// SymmetricKey is the base64-encoded 32-byte key for the current version.
	SymmetricKey string

	// KeyVersion identifies the current key version.
	KeyVersion string

	// AccessTTL is the time-to-live for access tokens. Defaults to DefaultAccessTTL.
	AccessTTL time.Duration

	// RefreshTTL is the time-to-live for refresh tokens. Defaults to DefaultRefreshTTL.
	RefreshTTL time.Duration
}

// NewKeyStore creates a new KeyStore from the given configuration.
func NewKeyStore(cfg KeyStoreConfig) (*KeyStore, error) {
	if cfg.SymmetricKey == "" {
		return nil, errors.New("paseto symmetric key is required")
	}

	if cfg.KeyVersion == "" {
		return nil, errors.New("paseto key version is required")
	}

	keyBytes, err := base64.StdEncoding.DecodeString(cfg.SymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode paseto key: %w", err)
	}

	if len(keyBytes) != keySize {
		return nil, fmt.Errorf("paseto key must be %d bytes, got %d", keySize, len(keyBytes))
	}

	key, err := paseto.V4SymmetricKeyFromBytes(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create paseto key: %w", err)
	}

	accessTTL := cfg.AccessTTL
	if accessTTL == 0 {
		accessTTL = DefaultAccessTTL
	}

	refreshTTL := cfg.RefreshTTL
	if refreshTTL == 0 {
		refreshTTL = DefaultRefreshTTL
	}

	return &KeyStore{
		currentKey:     key,
		currentVersion: cfg.KeyVersion,
		previousKeys:   make(map[string]paseto.V4SymmetricKey),
		accessTTL:      accessTTL,
		refreshTTL:     refreshTTL,
	}, nil
}

// GenerateTokenPair generates a new access/refresh token pair for the given claims.
func (ks *KeyStore) GenerateTokenPair(claims TokenClaims) (*TokenPair, error) {
	now := time.Now()

	accessToken := ks.generateToken(claims, "access", now, ks.accessTTL)
	refreshToken := ks.generateToken(claims, "refresh", now, ks.refreshTTL)

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(ks.accessTTL.Seconds()),
	}, nil
}

// ValidateToken validates a PASETO token and returns its claims.
// It tries the current key first, then falls back to previous keys.
func (ks *KeyStore) ValidateToken(token string) (*TokenClaims, error) {
	parsedToken, err := ks.parseToken(token)
	if err != nil {
		return nil, err
	}
	return extractClaims(parsedToken)
}

// ValidateRefreshToken validates a PASETO refresh token and returns its claims.
// Returns an error if the token is not a refresh token.
func (ks *KeyStore) ValidateRefreshToken(token string) (*TokenClaims, error) {
	claims, err := ks.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, errors.New("token is not a refresh token")
	}

	return claims, nil
}

// RefreshTokens validates a refresh token and generates a new token pair.
func (ks *KeyStore) RefreshTokens(refreshToken string) (*TokenPair, error) {
	claims, err := ks.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	return ks.GenerateTokenPair(TokenClaims{
		Subject: claims.Subject,
		Email:   claims.Email,
	})
}

// parseToken parses and validates a PASETO token using current and previous keys.
func (ks *KeyStore) parseToken(token string) (*paseto.Token, error) {
	parser := paseto.NewParser()
	parser.AddRule(paseto.ValidAt(time.Now()))

	parsedToken, err := parser.ParseV4Local(ks.currentKey, token, nil)
	if err != nil {
		for _, prevKey := range ks.previousKeys {
			parsedToken, err = parser.ParseV4Local(prevKey, token, nil)
			if err == nil {
				break
			}
		}
		if err != nil {
			return nil, fmt.Errorf("failed to validate paseto token: %w", err)
		}
	}

	return parsedToken, nil
}

// extractClaims extracts claims from a parsed PASETO token.
func extractClaims(token *paseto.Token) (*TokenClaims, error) {
	subject, err := token.GetSubject()
	if err != nil {
		return nil, fmt.Errorf("failed to get subject from token: %w", err)
	}

	email, _ := token.GetString("email")
	keyVersion, _ := token.GetString("key_version")
	tokenType, _ := token.GetString("token_type")

	return &TokenClaims{
		Subject:    subject,
		Email:      email,
		KeyVersion: keyVersion,
		TokenType:  tokenType,
	}, nil
}

// AddPreviousKey adds a previous key version for validation during key rotation.
func (ks *KeyStore) AddPreviousKey(version, encodedKey string) error {
	keyBytes, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return fmt.Errorf("failed to decode previous key: %w", err)
	}

	if len(keyBytes) != keySize {
		return fmt.Errorf("previous key must be %d bytes, got %d", keySize, len(keyBytes))
	}

	key, err := paseto.V4SymmetricKeyFromBytes(keyBytes)
	if err != nil {
		return fmt.Errorf("failed to create previous key: %w", err)
	}
	ks.previousKeys[version] = key

	return nil
}

func (ks *KeyStore) generateToken(
	claims TokenClaims, tokenType string, now time.Time, ttl time.Duration,
) string {
	token := paseto.NewToken()

	token.SetSubject(claims.Subject)
	token.SetString("email", claims.Email)
	token.SetString("key_version", ks.currentVersion)
	token.SetString("token_type", tokenType)

	token.SetIssuedAt(now)
	token.SetNotBefore(now)
	token.SetExpiration(now.Add(ttl))

	return token.V4Encrypt(ks.currentKey, nil)
}

// IsPASETOToken checks if a token string looks like a PASETO v4.local token.
func IsPASETOToken(token string) bool {
	return len(token) > len(TokenPrefix) && token[:len(TokenPrefix)] == TokenPrefix
}
