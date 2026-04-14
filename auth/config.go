package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	ExpiresAt  string // JWT 过期时间
	SigningKey string // HMAC 签名密钥
	RSAConf           // RSA 配置
}

type signingMethodOption struct {
	signature jwt.SigningMethod
	keyID     string
}

type Option func(c *signingMethodOption)

func WithSigningMethod(method SigningMethod) Option {
	return func(c *signingMethodOption) {
		c.signature = method
	}
}

func WithKeyID(keyID string) Option {
	return func(c *signingMethodOption) {
		c.keyID = keyID
	}
}
