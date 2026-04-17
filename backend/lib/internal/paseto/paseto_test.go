package paseto

import (
	"encoding/base64"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shaftoe/savetoink/backend/lib/consts"
)

func testKeyStore(t *testing.T) *KeyStore {
	t.Helper()
	key := paseto.NewV4SymmetricKey()
	encoded := base64.StdEncoding.EncodeToString(key.ExportBytes())

	ks, err := NewKeyStore(KeyStoreConfig{
		SymmetricKey: encoded,
		KeyVersion:   "v1",
		AccessTTL:    consts.DefaultAccessTTL,
		RefreshTTL:   consts.DefaultRefreshTTL,
	})
	require.NoError(t, err)
	return ks
}

func TestNewKeyStore(t *testing.T) {
	key := paseto.NewV4SymmetricKey()
	encoded := base64.StdEncoding.EncodeToString(key.ExportBytes())

	t.Run("valid configuration", func(t *testing.T) {
		ks, err := NewKeyStore(KeyStoreConfig{
			SymmetricKey: encoded,
			KeyVersion:   "v1",
		})
		require.NoError(t, err)
		assert.NotNil(t, ks)
		assert.Equal(t, "v1", ks.currentVersion)
		assert.Equal(t, consts.DefaultAccessTTL, ks.accessTTL)
		assert.Equal(t, consts.DefaultRefreshTTL, ks.refreshTTL)
	})

	t.Run("custom TTL", func(t *testing.T) {
		ks, err := NewKeyStore(KeyStoreConfig{
			SymmetricKey: encoded,
			KeyVersion:   "v1",
			AccessTTL:    30 * time.Minute,
			RefreshTTL:   24 * time.Hour,
		})
		require.NoError(t, err)
		assert.Equal(t, 30*time.Minute, ks.accessTTL)
		assert.Equal(t, 24*time.Hour, ks.refreshTTL)
	})

	t.Run("missing symmetric key", func(t *testing.T) {
		_, err := NewKeyStore(KeyStoreConfig{
			KeyVersion: "v1",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "symmetric key is required")
	})

	t.Run("missing key version", func(t *testing.T) {
		_, err := NewKeyStore(KeyStoreConfig{
			SymmetricKey: encoded,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "key version is required")
	})

	t.Run("invalid base64 key", func(t *testing.T) {
		_, err := NewKeyStore(KeyStoreConfig{
			SymmetricKey: "not-valid-base64!!!",
			KeyVersion:   "v1",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode paseto key")
	})

	t.Run("wrong key size", func(t *testing.T) {
		shortKey := base64.StdEncoding.EncodeToString([]byte("tooshort"))
		_, err := NewKeyStore(KeyStoreConfig{
			SymmetricKey: shortKey,
			KeyVersion:   "v1",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "paseto key must be 32 bytes")
	})
}

func TestGenerateTokenPair(t *testing.T) {
	ks := testKeyStore(t)
	claims := TokenClaims{
		Subject:    "auth0|123456",
		Email:      "user@example.com",
		KeyVersion: "v1",
	}

	t.Run("generates valid token pair", func(t *testing.T) {
		pair, err := ks.GenerateTokenPair(claims)
		require.NoError(t, err)

		assert.NotEmpty(t, pair.AccessToken)
		assert.NotEmpty(t, pair.RefreshToken)
		assert.Equal(t, int64(3600), pair.ExpiresIn) // 1 hour in seconds

		assert.True(t, IsPASETOToken(pair.AccessToken))
		assert.True(t, IsPASETOToken(pair.RefreshToken))
	})

	t.Run("access token contains correct claims", func(t *testing.T) {
		pair, err := ks.GenerateTokenPair(claims)
		require.NoError(t, err)

		validated, err := ks.ValidateToken(pair.AccessToken)
		require.NoError(t, err)

		assert.Equal(t, claims.Subject, validated.Subject)
		assert.Equal(t, claims.Email, validated.Email)
		assert.Equal(t, "v1", validated.KeyVersion)
		assert.Equal(t, "access", validated.TokenType)
	})

	t.Run("refresh token contains correct claims", func(t *testing.T) {
		pair, err := ks.GenerateTokenPair(claims)
		require.NoError(t, err)

		validated, err := ks.ValidateToken(pair.RefreshToken)
		require.NoError(t, err)

		assert.Equal(t, claims.Subject, validated.Subject)
		assert.Equal(t, claims.Email, validated.Email)
		assert.Equal(t, "v1", validated.KeyVersion)
		assert.Equal(t, "refresh", validated.TokenType)
	})

	t.Run("tokens are unique across calls", func(t *testing.T) {
		pair1, err := ks.GenerateTokenPair(claims)
		require.NoError(t, err)

		// Small sleep to ensure different nonce
		time.Sleep(10 * time.Millisecond)

		pair2, err := ks.GenerateTokenPair(claims)
		require.NoError(t, err)

		assert.NotEqual(t, pair1.AccessToken, pair2.AccessToken)
		assert.NotEqual(t, pair1.RefreshToken, pair2.RefreshToken)
	})
}

func TestValidateToken(t *testing.T) {
	ks := testKeyStore(t)
	claims := TokenClaims{
		Subject: "auth0|123456",
		Email:   "user@example.com",
	}

	t.Run("valid access token", func(t *testing.T) {
		pair, err := ks.GenerateTokenPair(claims)
		require.NoError(t, err)

		validated, err := ks.ValidateToken(pair.AccessToken)
		require.NoError(t, err)

		assert.Equal(t, claims.Subject, validated.Subject)
		assert.Equal(t, claims.Email, validated.Email)
		assert.Equal(t, "access", validated.TokenType)
	})

	t.Run("valid refresh token", func(t *testing.T) {
		pair, err := ks.GenerateTokenPair(claims)
		require.NoError(t, err)

		validated, err := ks.ValidateToken(pair.RefreshToken)
		require.NoError(t, err)

		assert.Equal(t, claims.Subject, validated.Subject)
		assert.Equal(t, "refresh", validated.TokenType)
	})

	t.Run("invalid token string", func(t *testing.T) {
		_, err := ks.ValidateToken("v4.local.invalidtoken")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to validate paseto token")
	})

	t.Run("empty token string", func(t *testing.T) {
		_, err := ks.ValidateToken("")
		require.Error(t, err)
	})

	t.Run("token encrypted with different key", func(t *testing.T) {
		otherKey := paseto.NewV4SymmetricKey()
		otherEncoded := base64.StdEncoding.EncodeToString(otherKey.ExportBytes())
		otherKS, err := NewKeyStore(KeyStoreConfig{
			SymmetricKey: otherEncoded,
			KeyVersion:   "other",
		})
		require.NoError(t, err)

		pair, err := otherKS.GenerateTokenPair(claims)
		require.NoError(t, err)

		_, err = ks.ValidateToken(pair.AccessToken)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to validate paseto token")
	})
}

func TestTokenExpiration(t *testing.T) {
	ks := testKeyStore(t)

	claims := TokenClaims{
		Subject: "auth0|123456",
		Email:   "user@example.com",
	}

	t.Run("freshly generated token is valid", func(t *testing.T) {
		pair, err := ks.GenerateTokenPair(claims)
		require.NoError(t, err)

		_, err = ks.ValidateToken(pair.AccessToken)
		require.NoError(t, err)

		_, err = ks.ValidateToken(pair.RefreshToken)
		require.NoError(t, err)
	})

	t.Run("expired token is rejected", func(t *testing.T) {
		// Manually create a token that expired 1 hour ago
		token := paseto.NewToken()
		token.SetSubject(claims.Subject)
		token.SetString("email", claims.Email)
		token.SetString("key_version", ks.currentVersion)
		token.SetString("token_type", "access")

		now := time.Now()
		token.SetIssuedAt(now.Add(-2 * time.Hour))
		token.SetNotBefore(now.Add(-2 * time.Hour))
		token.SetExpiration(now.Add(-1 * time.Hour))

		encrypted := token.V4Encrypt(ks.currentKey, nil)

		_, err := ks.ValidateToken(encrypted)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("not-yet-valid token is rejected", func(t *testing.T) {
		// Manually create a token that is not yet valid
		token := paseto.NewToken()
		token.SetSubject(claims.Subject)
		token.SetString("email", claims.Email)
		token.SetString("key_version", ks.currentVersion)
		token.SetString("token_type", "access")

		now := time.Now()
		token.SetIssuedAt(now)
		token.SetNotBefore(now.Add(1 * time.Hour))
		token.SetExpiration(now.Add(2 * time.Hour))

		encrypted := token.V4Encrypt(ks.currentKey, nil)

		_, err := ks.ValidateToken(encrypted)
		require.Error(t, err)
	})
}

func TestKeyRotation(t *testing.T) {
	oldKey := paseto.NewV4SymmetricKey()
	oldEncoded := base64.StdEncoding.EncodeToString(oldKey.ExportBytes())

	oldKS, err := NewKeyStore(KeyStoreConfig{
		SymmetricKey: oldEncoded,
		KeyVersion:   "v1",
	})
	require.NoError(t, err)

	claims := TokenClaims{
		Subject: "auth0|123456",
		Email:   "user@example.com",
	}

	// Generate tokens with old key
	pair, err := oldKS.GenerateTokenPair(claims)
	require.NoError(t, err)

	// Create new keystore with a new key
	newKey := paseto.NewV4SymmetricKey()
	newEncoded := base64.StdEncoding.EncodeToString(newKey.ExportBytes())

	newKS, err := NewKeyStore(KeyStoreConfig{
		SymmetricKey: newEncoded,
		KeyVersion:   "v2",
	})
	require.NoError(t, err)

	t.Run("old tokens fail without previous key", func(t *testing.T) {
		_, validateErr := newKS.ValidateToken(pair.AccessToken)
		require.Error(t, validateErr)
	})

	t.Run("old tokens validate after adding previous key", func(t *testing.T) {
		addErr := newKS.AddPreviousKey("v1", oldEncoded)
		require.NoError(t, addErr)

		validated, validateErr := newKS.ValidateToken(pair.AccessToken)
		require.NoError(t, validateErr)
		assert.Equal(t, claims.Subject, validated.Subject)
		assert.Equal(t, "v1", validated.KeyVersion)
	})

	t.Run("new tokens validate with current key", func(t *testing.T) {
		newPair, genErr := newKS.GenerateTokenPair(claims)
		require.NoError(t, genErr)

		validated, validateErr := newKS.ValidateToken(newPair.AccessToken)
		require.NoError(t, validateErr)
		assert.Equal(t, claims.Subject, validated.Subject)
		assert.Equal(t, "v2", validated.KeyVersion)
	})
}

func TestAddPreviousKey(t *testing.T) {
	ks := testKeyStore(t)

	t.Run("valid previous key", func(t *testing.T) {
		key := paseto.NewV4SymmetricKey()
		encoded := base64.StdEncoding.EncodeToString(key.ExportBytes())

		err := ks.AddPreviousKey("v0", encoded)
		require.NoError(t, err)
		assert.Contains(t, ks.previousKeys, "v0")
	})

	t.Run("invalid base64", func(t *testing.T) {
		err := ks.AddPreviousKey("v0", "!!!invalid!!!")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode previous key")
	})

	t.Run("wrong size key", func(t *testing.T) {
		shortKey := base64.StdEncoding.EncodeToString([]byte("tooshort"))
		err := ks.AddPreviousKey("v0", shortKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "previous key must be 32 bytes")
	})
}

func TestIsPASETOToken(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid v4.local token",
			input: "v4.local.someencryptedpayload",
			want:  true,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
		{
			name:  "JWT-like token",
			input: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.payload.signature",
			want:  false,
		},
		{
			name:  "v4.public token",
			input: "v4.public.somesignedpayload",
			want:  false,
		},
		{
			name:  "just the prefix",
			input: "v4.local.",
			want:  false,
		},
		{
			name:  "prefix only",
			input: "v4.local",
			want:  false,
		},
		{
			name:  "bearer prefix",
			input: "Bearer v4.local.something",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsPASETOToken(tt.input))
		})
	}
}

func TestKeyStore_SubjectPreservation(t *testing.T) {
	ks := testKeyStore(t)

	tests := []struct {
		name    string
		subject string
		email   string
	}{
		{
			name:    "standard Auth0 subject",
			subject: "auth0|123456789",
			email:   "user@example.com",
		},
		{
			name:    "google OAuth subject",
			subject: "google-oauth2|10987654321",
			email:   "someone@gmail.com",
		},
		{
			name:    "subject with special characters",
			subject: "auth0|abc-def_ghi",
			email:   "user+tag@example.com",
		},
		{
			name:    "long email",
			subject: "auth0|999",
			email:   "very.long.email.address.that.is.quite.long@subdomain.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := TokenClaims{
				Subject: tt.subject,
				Email:   tt.email,
			}

			pair, err := ks.GenerateTokenPair(claims)
			require.NoError(t, err)

			validated, err := ks.ValidateToken(pair.AccessToken)
			require.NoError(t, err)

			assert.Equal(t, tt.subject, validated.Subject)
			assert.Equal(t, tt.email, validated.Email)
		})
	}
}

func TestValidateRefreshToken(t *testing.T) {
	ks := testKeyStore(t)
	claims := TokenClaims{
		Subject: "auth0|123456",
		Email:   "user@example.com",
	}

	t.Run("valid refresh token", func(t *testing.T) {
		pair, err := ks.GenerateTokenPair(claims)
		require.NoError(t, err)

		validated, err := ks.ValidateRefreshToken(pair.RefreshToken)
		require.NoError(t, err)
		assert.Equal(t, claims.Subject, validated.Subject)
		assert.Equal(t, "refresh", validated.TokenType)
	})

	t.Run("access token rejected as refresh token", func(t *testing.T) {
		pair, err := ks.GenerateTokenPair(claims)
		require.NoError(t, err)

		_, err = ks.ValidateRefreshToken(pair.AccessToken)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a refresh token")
	})

	t.Run("invalid token rejected", func(t *testing.T) {
		_, err := ks.ValidateRefreshToken("v4.local.invalid")
		require.Error(t, err)
	})
}

func TestRefreshTokens(t *testing.T) {
	ks := testKeyStore(t)
	claims := TokenClaims{
		Subject: "auth0|123456",
		Email:   "user@example.com",
	}

	t.Run("valid refresh returns new token pair", func(t *testing.T) {
		originalPair, err := ks.GenerateTokenPair(claims)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		newPair, err := ks.RefreshTokens(originalPair.RefreshToken)
		require.NoError(t, err)

		assert.NotEqual(t, originalPair.AccessToken, newPair.AccessToken)
		assert.NotEqual(t, originalPair.RefreshToken, newPair.RefreshToken)

		validated, validateErr := ks.ValidateToken(newPair.AccessToken)
		require.NoError(t, validateErr)
		assert.Equal(t, claims.Subject, validated.Subject)
		assert.Equal(t, claims.Email, validated.Email)
		assert.Equal(t, "access", validated.TokenType)

		refreshValidated, refreshErr := ks.ValidateRefreshToken(newPair.RefreshToken)
		require.NoError(t, refreshErr)
		assert.Equal(t, claims.Subject, refreshValidated.Subject)
	})

	t.Run("access token cannot be used to refresh", func(t *testing.T) {
		pair, err := ks.GenerateTokenPair(claims)
		require.NoError(t, err)

		_, err = ks.RefreshTokens(pair.AccessToken)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid refresh token")
	})

	t.Run("invalid token rejected", func(t *testing.T) {
		_, err := ks.RefreshTokens("v4.local.invalid")
		require.Error(t, err)
	})

	t.Run("refreshed refresh token can be used again", func(t *testing.T) {
		pair1, err := ks.GenerateTokenPair(claims)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		pair2, err := ks.RefreshTokens(pair1.RefreshToken)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		pair3, err := ks.RefreshTokens(pair2.RefreshToken)
		require.NoError(t, err)

		validated, validateErr := ks.ValidateToken(pair3.AccessToken)
		require.NoError(t, validateErr)
		assert.Equal(t, claims.Subject, validated.Subject)
	})
}
