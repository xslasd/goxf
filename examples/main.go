package main

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xslasd/goxf"
	"github.com/xslasd/goxf/auth"
	"github.com/xslasd/goxf/ecode"
	"github.com/xslasd/goxf/i18n"
	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/server/sgin"
	"github.com/xslasd/goxf/utils/xfmt"
)

type MyClaims struct {
	jwt.RegisteredClaims
}

func main() {
	srv := goxf.NewService(
		goxf.WithConfigPassword("123456"),
	)

	lang, err := i18n.NewI18N()
	if err != nil {
		panic(err)
	}
	lang_en := lang.Language("en")
	xfmt.Printf("demo.test:%s", lang_en.T("demo.test"))
	xfmt.Printf("user.name:%s", lang.T("user.name"))

	certifier, err := auth.NewJWTCertifier()
	if err != nil {
		fmt.Println(ecode.As(err).Message(), "--")
		panic(err)
	}
	claims := &MyClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "dddd",
			ExpiresAt: certifier.NextExpiresAt(),
		},
	}

	token, err := certifier.GenerateToken(claims)

	if err != nil {
		panic(err.Error())
	}

	claimsData, err := certifier.ParseToken(token)
	if err != nil {
		fmt.Println(ecode.As(err).Message())
		panic(err)
	}
	xfmt.Printf("data:%v", claimsData)

	ginSrv, err := sgin.NewGinServer()
	if err != nil {
		panic(err)
	}

	if err := srv.Run(ginSrv); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
