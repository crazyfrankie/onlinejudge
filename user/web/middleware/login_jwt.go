package middleware

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrTokenInvalid = errors.New("token is invalid")
	ErrTokenExpired = errors.New("token is expired")
	ErrLoginYet     = errors.New("have not logged in yet")
	SecretKey       = []byte("KsS2X1CgFT4bi3BRRIxLk5jjiUBj8wxE")
)

type Claims struct {
	Role      uint8  `json:"role"`
	Id        uint64 `json:"id"`
	UserAgent string `json:"userAgent"`
	jwt.StandardClaims
}

// LoginJWTMiddlewareBuilder JWT 进行校验
type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{paths: make([]string, 0)}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(paths string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, paths)
	return l
}

func (l *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 路径校验
		for _, path := range l.paths {
			if c.Request.URL.Path == path {
				return
			}
		}

		tokenHeader := c.GetHeader("Authorization")

		// 检查请求头中是否包含 Token
		if tokenHeader == "" {
			// 没登录
			c.JSON(http.StatusUnauthorized, "you need to login")
			c.Abort()
			return
		}

		claims, err := l.ParseToken(tokenHeader)
		if err != nil {
			code, msg := handleTokenError(err)
			c.JSON(code, msg)
			c.Abort()
			return
		}

		if claims.UserAgent != c.Request.UserAgent() {
			// 严重的安全问题
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

func (l *LoginJWTMiddlewareBuilder) ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})
	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, ErrTokenInvalid
			} else if ve.Errors&(jwt.ValidationErrorExpired) != 0 {
				return nil, ErrTokenExpired
			} else if ve.Errors&(jwt.ValidationErrorNotValidYet) != 0 {
				return nil, ErrLoginYet
			}
		}
		return nil, err
	}
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, ErrTokenInvalid
}

func handleTokenError(err error) (int, string) {
	var code int
	var msg string
	switch {
	case errors.Is(err, ErrTokenExpired):
		code = http.StatusUnauthorized
		msg = "token is expired"
	case errors.Is(err, ErrTokenInvalid):
		code = http.StatusUnauthorized
		msg = "token is invalid"
	case errors.Is(err, ErrLoginYet):
		code = http.StatusUnauthorized
		msg = "have not logged in yet"
	default:
		code = http.StatusInternalServerError
		msg = "parse Token failed"
	}
	return code, msg
}
