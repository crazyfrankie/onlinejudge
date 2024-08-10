package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

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

		_, err := ParseToken(tokenHeader)
		if err != nil {
			code, msg := handleTokenError(err)
			c.JSON(code, msg)
			c.Abort()
			return
		}

		c.Next()
	}
}
