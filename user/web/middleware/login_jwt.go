package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// 创建者模式

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
		token := c.GetHeader("Authorization")

		// 检查请求头中是否包含 Token
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": http.StatusNotFound,
				"msg":    "require parameter error",
				"data":   "lack of Token",
			})

			c.Abort()
			return
		}

		_, err := ParseToken(token)
		if err != nil {
			code, msg := handleTokenError(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": code,
				"msg":    "operate failed",
				"data":   msg,
			})
			c.Abort()
			return
		}
	}
}