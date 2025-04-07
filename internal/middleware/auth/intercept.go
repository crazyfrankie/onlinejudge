package auth

import (
	"errors"
	"github.com/crazyfrankie/onlinejudge/common/constant"
	er "github.com/crazyfrankie/onlinejudge/common/errors"
	"github.com/crazyfrankie/onlinejudge/common/response"
	jwt3 "github.com/crazyfrankie/onlinejudge/internal/middleware/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/redis/go-redis/v9"
)

var (
	ErrTokenInvalid = errors.New("token is invalid")
	ErrTokenExpired = errors.New("token is expired")
	ErrLoginYet     = errors.New("have not logged in yet")
)

// LoginJWTMiddlewareBuilder JWT 进行校验
type LoginJWTMiddlewareBuilder struct {
	ignorePaths map[string]struct{}
	adminPaths  map[string]struct{}
	jwt3.Handler
	cmd redis.Cmdable
}

func NewLoginJWTMiddlewareBuilder(jwtHdl jwt3.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		ignorePaths: make(map[string]struct{}),
		adminPaths:  make(map[string]struct{}),
		Handler:     jwtHdl,
	}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.ignorePaths[path] = struct{}{}
	return l
}

func (l *LoginJWTMiddlewareBuilder) AdminPaths(path string) *LoginJWTMiddlewareBuilder {
	l.adminPaths[path] = struct{}{}
	return l
}

func (l *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 路径校验
		if _, ok := l.ignorePaths[c.Request.URL.Path]; ok {
			c.Next()
			return
		}

		// 提取并检查 token
		token := l.Handler.ExtractToken(c)

		claims, err := ParseToken(token)
		if err != nil {
			errCode := handleTokenError(err)
			response.Error(c, errCode)
			return
		}

		if claims.UserAgent != c.Request.UserAgent() {
			// 严重的安全问题
			response.Error(c, er.NewBizError(constant.ErrUserUnauthorized))
			return
		}

		err = l.Handler.CheckSession(c, claims.SSId)
		if err != nil {
			response.Error(c, er.NewBizError(constant.ErrUserSessExpired))
			return
		}

		if _, ok := l.adminPaths[c.Request.URL.Path]; ok && claims.Role != 1 {
			response.Error(c, er.NewBizError(constant.ErrUserForbidden))
			return
		}

		// 将解析出来的 Claims 存入上下文
		c.Set("claims", claims)
		// 继续后续的处理
		c.Next()
	}
}

func ParseToken(token string) (*jwt3.Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &jwt3.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwt3.SecretKey, nil
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
		if claims, ok := tokenClaims.Claims.(*jwt3.Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, ErrTokenInvalid
}

func handleTokenError(err error) *er.BizError {
	var errCode constant.ErrorCode

	switch {
	case errors.Is(err, ErrTokenInvalid):
		errCode = constant.ErrUserInvalidToken
	case errors.Is(err, ErrTokenExpired):
		errCode = constant.ErrUserTokenExpired
	case errors.Is(err, ErrLoginYet):
		errCode = constant.ErrUserLoginYet
	default:
		errCode = constant.ErrUserInternalServer
	}

	return er.NewBizError(errCode)
}
