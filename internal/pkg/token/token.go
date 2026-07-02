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

var ErrInvalidToken = kratoserrors.Unauthorized("UNAUTHORIZED", "invalid token")

// AccessTokenClaims 表示访问令牌中的业务声明。
type AccessTokenClaims struct {
	UserID                 string `json:"uid"`     // UserID 是用户编号。
	PairID                 string `json:"pair_id"` // PairID 是访问令牌和刷新令牌的配对编号。
	JTI                    string `json:"jti"`     // JTI 是访问令牌唯一编号。
	Version                string `json:"ver"`     // Version 是用户 token 版本。
	Type                   string `json:"typ"`     // Type 是令牌类型。
	jwtv5.RegisteredClaims        // RegisteredClaims 是 JWT 标准声明。
}

// RefreshTokenClaims 表示刷新令牌中的业务声明。
type RefreshTokenClaims struct {
	UserID                 string `json:"uid"` // UserID 是用户编号。
	PairID                 string `json:"jti"` // PairID 是令牌配对编号。
	Version                string `json:"ver"` // Version 是用户 token 版本。
	Type                   string `json:"typ"` // Type 是令牌类型。
	jwtv5.RegisteredClaims        // RegisteredClaims 是 JWT 标准声明。
}

// Pair 表示访问令牌和刷新令牌组合。
type Pair struct {
	AccessToken  string `json:"access_token"`  // AccessToken 是访问令牌。
	RefreshToken string `json:"refresh_token"` // RefreshToken 是刷新令牌。
	ExpiresIn    int64  `json:"expires_in"`    // ExpiresIn 是访问令牌有效秒数。
}

// Manager 管理 Ed25519 JWT 令牌签发和校验。
type Manager struct {
	privateKey      ed25519.PrivateKey // privateKey 是令牌签名私钥。
	publicKey       ed25519.PublicKey  // publicKey 是令牌验签公钥。
	accessTokenTTL  time.Duration      // accessTokenTTL 是访问令牌有效期。
	refreshTokenTTL time.Duration      // refreshTokenTTL 是刷新令牌有效期。
}

// NewManager 创建令牌管理器。
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

// GenerateTokenPair 为用户签发访问令牌和刷新令牌。
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
	claims := &RefreshTokenClaims{
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
	refreshToken, err := jwtv5.NewWithClaims(jwtv5.SigningMethodEdDSA, claims).SignedString(m.privateKey)
	if err != nil {
		return nil, err
	}
	return &Pair{AccessToken: accessToken, RefreshToken: refreshToken, ExpiresIn: int64(m.accessTokenTTL.Seconds())}, nil
}

// GenerateAccessToken 为指定令牌配对签发访问令牌。
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

// ValidateAccessToken 校验访问令牌并返回声明。
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

// loadOrGeneratePrivateKey 读取 Ed25519 私钥或生成临时私钥。
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

// randomID 生成随机十六进制编号。
func randomID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
