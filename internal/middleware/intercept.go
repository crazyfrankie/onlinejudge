package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/redis/go-redis/v9"

	ijwt "oj/internal/user/web/jwt"
)

var (
	ErrTokenInvalid = errors.New("token is invalid")
	ErrTokenExpired = errors.New("token is expired")
	ErrLoginYet     = errors.New("have not logged in yet")
)

// LoginJWTMiddlewareBuilder JWT 进行校验
type LoginJWTMiddlewareBuilder struct {
	paths map[string]struct{}
	ijwt.Handler
	cmd redis.Cmdable
}

func NewLoginJWTMiddlewareBuilder(jwtHdl ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		paths:   make(map[string]struct{}),
		Handler: jwtHdl,
	}
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
		tokenHeader := l.Handler.ExtractToken(c)

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

		err = l.Handler.CheckSession(c, claims.SSId)
		if err != nil {
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
	ijwt.Handler
	cmd redis.Cmdable
}

func NewProblemJWTMiddlewareBuilder(jwtHdl ijwt.Handler) *ProblemJWTMiddlewareBuilder {
	return &ProblemJWTMiddlewareBuilder{
		paths:   make(map[string]struct{}),
		Handler: jwtHdl,
	}
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

		tokenHeader := l.Handler.ExtractToken(c)

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

		err = l.Handler.CheckSession(c, claims.SSId)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 将解析出来的 Claims 存入上下文
		c.Set("claims", claims)
		// 继续后续的处理
		c.Next()
	}
}

func ParseToken(token string) (*ijwt.Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &ijwt.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return ijwt.SecretKey, nil
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
		if claims, ok := tokenClaims.Claims.(*ijwt.Claims); ok && tokenClaims.Valid {
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
