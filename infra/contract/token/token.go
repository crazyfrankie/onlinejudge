package token

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	er "github.com/crazyfrankie/onlinejudge/common/errors"
)

type Token interface {
	SetLoginToken(ctx *gin.Context, uid uint64) error
	AccessToken(ctx *gin.Context, id uint64, ssid string) (string, error)
	RefreshToken(ctx *gin.Context, id uint64, ssid string) (string, error)
	ParseToken(token string) (*Claims, error)
	CheckSession(ctx *gin.Context, uid uint64, ssid string) error
	ClearToken(ctx *gin.Context) error
	TryRefresh(ctx *gin.Context) error
	LogoutAllDevices(ctx *gin.Context) error
	HandleTokenError(err error) *er.BizError
}

type Claims struct {
	Id        uint64 `json:"id"`
	UserAgent string `json:"userAgent"`
	SSId      string
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	Role      uint8
	Id        uint64
	UserAgent string
	SSId      string
	jwt.RegisteredClaims
}
