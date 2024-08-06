package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// 创建者模式

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{paths: make([]string, 0)}
}

func (l *LoginMiddlewareBuilder) IgnorePaths(paths string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, paths)
	return l
}

func (l *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 不需要登录校验的
		for _, path := range l.paths {
			if c.Request.URL.Path == path {
				return
			}
		}

		sess := sessions.Default(c)
		if sess.Get("identifier") == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

	}
}
