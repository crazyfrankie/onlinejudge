package middleware

import (
	"github.com/gin-gonic/contrib/cors"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

func Cors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		// 是否允许带 cookie 之类的
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "localhost") {
				// 开发环境
				return true
			}
			return strings.Contains(origin, "yourcompany.com")
		},
		// 不加这一行 前端拿不到 token
		ExposedHeaders: []string{"x-jwt-token"},
		MaxAge:         12 * time.Hour,
	})
}
