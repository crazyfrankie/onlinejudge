package jwt

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/golang-jwt/jwt"
)

type Handler interface {
	SetLoginToken(ctx *app.RequestContext, role uint8, uid uint64) error
	AccessToken(ctx *app.RequestContext, role uint8, id uint64, ssid string) error
	RefreshToken(ctx *app.RequestContext, role uint8, id uint64, ssid string) error
	ExtractToken(ctx *app.RequestContext) string
	CheckSession(ctx *app.RequestContext, ssid string) error
	ClearToken(ctx *app.RequestContext) error
}

type Claims struct {
	Role      uint8  `json:"role"`
	Id        uint64 `json:"id"`
	UserAgent string `json:"userAgent"`
	SSId      string
	jwt.StandardClaims
}

type RefreshClaims struct {
	Role      uint8
	Id        uint64
	UserAgent string
	SSId      string
	jwt.StandardClaims
}
