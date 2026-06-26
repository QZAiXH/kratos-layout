package token

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

const (
	TypeAccess  = "access"
	TypeRefresh = "refresh"
)

var (
	ErrInvalidToken = kratoserrors.Unauthorized("UNAUTHORIZED", "invalid token")
)

type AccessTokenClaims struct {
	UserID  string `json:"uid"`
	PairID  string `json:"pair_id"`
	JTI     string `json:"jti"`
	Version string `json:"ver"`
	Type    string `json:"typ"`
	jwtv5.RegisteredClaims
}

type RefreshTokenClaims struct {
	UserID  string `json:"uid"`
	PairID  string `json:"jti"`
	Version string `json:"ver"`
	Type    string `json:"typ"`
	jwtv5.RegisteredClaims
}

type Pair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type Manager struct {
	privateKey      ed25519.PrivateKey
	publicKey       ed25519.PublicKey
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewManager(privateKeyPath string, accessTTL, refreshTTL time.Duration) (*Manager, error) {
	if accessTTL <= 0 {
		accessTTL = 2 * time.Hour
	}
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}

	privateKey, err := loadOrGeneratePrivateKey(privateKeyPath)
	if err != nil {
		return nil, err
	}

	return &Manager{
		privateKey:      privateKey,
		publicKey:       privateKey.Public().(ed25519.PublicKey),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}, nil
}

func (m *Manager) GenerateTokenPair(userID, tokenVersion string) (*Pair, error) {
	pairID, err := randomID()
	if err != nil {
		return nil, err
	}
	accessToken, err := m.GenerateAccessToken(userID, tokenVersion, pairID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	refreshClaims := &RefreshTokenClaims{
		UserID:  userID,
		PairID:  pairID,
		Version: tokenVersion,
		Type:    TypeRefresh,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(now.Add(m.refreshTokenTTL)),
			IssuedAt:  jwtv5.NewNumericDate(now),
			NotBefore: jwtv5.NewNumericDate(now),
		},
	}
	refreshToken, err := jwtv5.NewWithClaims(jwtv5.SigningMethodEdDSA, refreshClaims).SignedString(m.privateKey)
	if err != nil {
		return nil, err
	}
	return &Pair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(m.accessTokenTTL.Seconds()),
	}, nil
}

func (m *Manager) GenerateAccessToken(userID, tokenVersion, pairID string) (string, error) {
	jti, err := randomID()
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := &AccessTokenClaims{
		UserID:  userID,
		PairID:  pairID,
		JTI:     jti,
		Version: tokenVersion,
		Type:    TypeAccess,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(now.Add(m.accessTokenTTL)),
			IssuedAt:  jwtv5.NewNumericDate(now),
			NotBefore: jwtv5.NewNumericDate(now),
		},
	}
	return jwtv5.NewWithClaims(jwtv5.SigningMethodEdDSA, claims).SignedString(m.privateKey)
}

func (m *Manager) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	parsed, err := jwtv5.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwtv5.Token) (any, error) {
		if _, ok := token.Method.(*jwtv5.SigningMethodEd25519); !ok {
			return nil, ErrInvalidToken
		}
		return m.publicKey, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(*AccessTokenClaims)
	if !ok || !parsed.Valid || claims.Type != TypeAccess || strings.TrimSpace(claims.UserID) == "" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func loadOrGeneratePrivateKey(privateKeyPath string) (ed25519.PrivateKey, error) {
	privateKeyPath = strings.TrimSpace(privateKeyPath)
	if privateKeyPath == "" {
		_, privateKey, err := ed25519.GenerateKey(rand.Reader)
		return privateKey, err
	}

	keyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("decode private key PEM")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	ed25519Key, ok := privateKey.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not Ed25519")
	}
	return ed25519Key, nil
}

func randomID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
