package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// LoginJWTMiddlewareBuilder JWT 进行校验
type LoginJWTMiddlewareBuilder struct {
	paths map[string]struct{}
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{paths: make(map[string]struct{})}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths[path] = struct{}{}
	return l
}

func (l *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 路径校验
		if _, ok := l.paths[c.Request.URL.Path]; ok {
			c.Next()
			return
		}

		// 提取并检查 token
		tokenHeader := ExtractToken(c)

		claims, err := ParseToken(tokenHeader)
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

		// 将解析出来的 Claims 存入上下文
		c.Set("claims", claims)
		// 继续后续的处理
		c.Next()
	}
}

type ProblemJWTMiddlewareBuilder struct {
	paths map[string]struct{}
}

func NewProblemJWTMiddlewareBuilder() *ProblemJWTMiddlewareBuilder {
	return &ProblemJWTMiddlewareBuilder{paths: make(map[string]struct{})}
}

func (l *ProblemJWTMiddlewareBuilder) SecretPaths(path string) *ProblemJWTMiddlewareBuilder {
	l.paths[path] = struct{}{}
	return l
}

func (l *ProblemJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := l.paths[c.Request.URL.Path]; !ok {
			c.Next()
			return
		}

		tokenHeader := c.GetHeader("Authorization")

		// 检查请求头中是否包含 Token
		if tokenHeader == "" {
			// 没登录
			c.JSON(http.StatusUnauthorized, "you need to login")
			c.Abort()
			return
		}

		claims, err := ParseToken(tokenHeader)
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

		if claims.Role != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, "access denied")
			return
		}

		// 将解析出来的 Claims 存入上下文
		c.Set("claims", claims)
		// 继续后续的处理
		c.Next()
	}
}

func ExtractToken(ctx *gin.Context) string {
	tokenHeader := ctx.GetHeader("Authorization")
	// 检查请求头中是否包含 Token
	if tokenHeader == "" {
		return ""
	}
	return tokenHeader
}
