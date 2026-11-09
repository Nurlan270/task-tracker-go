package jwt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	errs "github.com/Nurlan270/task-tracker-go/auth-service/internal/errors"
)

var signingMethod = jwt.SigningMethodHS256

type Claims = jwt.MapClaims

type AccessToken struct {
	Value     string
	ExpiresIn time.Duration
}

type RefreshToken struct {
	Value     string
	Hash      string
	ExpiresAt time.Time
}

type TokenPair struct {
	AccessToken  AccessToken
	RefreshToken RefreshToken
}

type TokenManager struct {
	cfg Config
}

func NewTokenManager(cfg Config) *TokenManager {
	return &TokenManager{cfg: cfg}
}

// Generate generates token pair containing provided claims and refresh token.
func (m *TokenManager) Generate(c Claims) (TokenPair, error) {
	//	default claims
	claims := jwt.MapClaims{
		"exp": time.Now().Add(m.cfg.AccessTTL).Unix(),
		"nbf": time.Now().Unix(),
		"iat": time.Now().Unix(),
		"iss": m.cfg.Issuer,
	}

	//	add custom claims
	for k, v := range c {
		claims[k] = v
	}

	token := jwt.NewWithClaims(signingMethod, claims)

	accessToken, err := token.SignedString([]byte(m.cfg.Secret))
	if err != nil {
		return TokenPair{}, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return TokenPair{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return TokenPair{
		AccessToken: AccessToken{
			Value:     accessToken,
			ExpiresIn: m.cfg.AccessTTL,
		},
		RefreshToken: RefreshToken{
			Value:     refreshToken,
			Hash:      m.HashToken(refreshToken),
			ExpiresAt: time.Now().Add(m.cfg.RefreshTTL),
		},
	}, nil
}

// Parse parses provided tokenString and fetches its claims.
func (m *TokenManager) Parse(tokenString string) (Claims, error) {
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{signingMethod.Alg()}),
		jwt.WithIssuer(m.cfg.Issuer),
		jwt.WithExpirationRequired(),
		jwt.WithNotBeforeRequired(),
		jwt.WithIssuedAt(),
	}

	keyFunc := func(token *jwt.Token) (any, error) {
		return []byte(m.cfg.Secret), nil
	}

	token, err := jwt.Parse(tokenString, keyFunc, opts...)
	if err != nil {
		return nil, errs.ErrTokenInvalid
	} else if !token.Valid {
		return nil, errs.ErrTokenInvalid
	}

	claims, ok := token.Claims.(Claims)
	if !ok {
		return nil, errors.Join(errs.ErrInternal, errors.New("failed to convert claims"))
	}

	return claims, nil
}

// HashToken hashes provided token using SHA256 hashing algorithm.
func (m *TokenManager) HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func generateRefreshToken() (string, error) {
	// 32 random bytes = 256 bits of entropy.
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
