package auth

import (
	"time"

	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/conf"
	"github.com/xslasd/goxf/ecode"
	"github.com/xslasd/goxf/log"
)

func NewJWTCertifier(options ...Option) (Certifier, error) {
	application.CheckStartupGoxf()
	cfg := &Config{
		ExpiresAt: "2h",
	}
	key := "jwt"
	if err := conf.UnmarshalKey(key, cfg); err != nil {
		return nil, err
	}
	expiresAt, err := time.ParseDuration(cfg.ExpiresAt)
	if err != nil {
		log.Panic("jwt expiresAt 解析错误：", log.FieldErr(err))
	}
	if cfg.SigningKey != "" {
		return NewHMAC(cfg.SigningKey, expiresAt, options...)
	}
	if cfg.RSAConf.PublicFile != "" || cfg.RSAConf.PrivateFile != "" {
		return NewRSA(cfg.RSAConf, expiresAt, options...)
	}
	return nil, ecode.SigningKeyIsNull
}
