package auth

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xslasd/goxf/ecode"
)

type SigningMethod = jwt.SigningMethod

var (
	SigningMethodHS256 = jwt.SigningMethodHS256
	SigningMethodHS384 = jwt.SigningMethodHS384
	SigningMethodHS512 = jwt.SigningMethodHS512

	SigningMethodRS256 = jwt.SigningMethodRS256
	SigningMethodRS384 = jwt.SigningMethodRS384
	SigningMethodRS512 = jwt.SigningMethodRS512
)

type Certifier interface {
	// GenerateToken 生成Token
	GenerateToken(claims jwt.Claims) (string, error)
	// ParseToken 解析Token
	ParseToken(token string) (jwt.Claims, error)
	// RefreshToken 刷新Token
	RefreshToken(claims jwt.Claims) (string, error)
	// NextExpiresAt 获取下一个Token过期时间
	NextExpiresAt() *jwt.NumericDate
}

type authCertifier struct {
	signedKey any // 签名密钥
	parserKey any // 解析密钥

	expiresAt time.Duration // 过期时间

	signature signingMethodOption // 签名配置
}

func NewHMAC(signingKey string, expiresAt time.Duration, options ...Option) (Certifier, error) {
	if signingKey == "" {
		return nil, fmt.Errorf("HMAC signingKey密钥不能为空")
	}
	keyBytes := []byte(signingKey)
	// 安全建议：HMAC密钥应至少32字节（256位）
	if len(keyBytes) < 32 {
		return nil, ecode.SigningKeyLimit.SetMessagef(strconv.Itoa(len(keyBytes)))
	}
	data := new(authCertifier)
	data.signedKey = keyBytes
	data.parserKey = keyBytes
	data.expiresAt = expiresAt
	data.signature = signingMethodOption{signature: SigningMethodHS256}
	for _, o := range options {
		o(&data.signature)
	}
	return data, nil
}

type RSAConf struct {
	PrivateFile string // 私钥文件
	PublicFile  string // 公钥文件
}

// NewRSA 创建RSA实例
// key: RSA密钥对
// options: 选项
func NewRSA(key RSAConf, expiresAt time.Duration, options ...Option) (Certifier, error) {
	data := new(authCertifier)
	if key.PublicFile != "" {
		publicKey, err := os.ReadFile(key.PublicFile)
		if err != nil {
			return nil, err
		}
		// 预解析公钥
		data.parserKey, err = jwt.ParseRSAPublicKeyFromPEM(publicKey)
		if err != nil {
			return nil, err
		}
	}
	if key.PrivateFile != "" {
		privateKey, err := os.ReadFile(key.PrivateFile)
		if err != nil {
			return nil, err
		}
		// 预解析私钥
		data.signedKey, err = jwt.ParseRSAPrivateKeyFromPEM(privateKey)
		if err != nil {
			return nil, err
		}
	}
	if data.parserKey == nil || data.signedKey == nil {
		return nil, fmt.Errorf("RSA 密钥对不能为空")
	}
	data.expiresAt = expiresAt
	data.signature = signingMethodOption{signature: SigningMethodRS256}
	for _, o := range options {
		o(&data.signature)
	}
	return data, nil
}

func (a authCertifier) GenerateToken(claims jwt.Claims) (string, error) {
	tokenClaims := jwt.NewWithClaims(a.signature.signature, claims)
	if a.signature.keyID != "" {
		tokenClaims.Header["kid"] = a.signature.keyID
	}
	return tokenClaims.SignedString(a.signedKey)
}

func (a authCertifier) ParseToken(token string) (jwt.Claims, error) {
	tokenClaims, err := jwt.Parse(token, a.keyFunc)
	if err != nil {
		return nil, err
	}
	if !tokenClaims.Valid {
		return nil, ecode.AccessTokenInvalid
	}
	return tokenClaims.Claims, nil
}

func (a authCertifier) keyFunc(token *jwt.Token) (any, error) {
	alg := token.Method.Alg()
	myAlg := a.signature.signature.Alg()
	if alg != myAlg {
		return nil, fmt.Errorf("签名方法不匹配: 期望 %s, 实际 %s",
			myAlg, alg)
	}
	if a.signature.keyID != "" && token.Header["kid"] != a.signature.keyID {
		return nil, fmt.Errorf("密钥kid 不匹配")
	}
	return a.parserKey, nil
}

func (a authCertifier) RefreshToken(claims jwt.Claims) (string, error) {
	return a.GenerateToken(claims)
}

func (a authCertifier) NextExpiresAt() *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(a.expiresAt))
}
