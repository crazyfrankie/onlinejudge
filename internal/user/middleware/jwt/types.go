package jwt

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type Handler interface {
	SetLoginToken(ctx *gin.Context, role uint8, uid uint64) error
	AccessToken(ctx *gin.Context, role uint8, id uint64, ssid string) error
	RefreshToken(ctx *gin.Context, role uint8, id uint64, ssid string) error
	ExtractToken(ctx *gin.Context) string
	CheckSession(ctx *gin.Context, ssid string) error
	ClearToken(ctx *gin.Context) error
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
