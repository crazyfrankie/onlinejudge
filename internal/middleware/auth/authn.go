package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	sjwt "github.com/golang-jwt/jwt"
	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/onlinejudge/common/constant"
	er "github.com/crazyfrankie/onlinejudge/common/errors"
	"github.com/crazyfrankie/onlinejudge/common/response"
	"github.com/crazyfrankie/onlinejudge/internal/middleware/jwt"
)

var (
	ErrTokenInvalid = errors.New("token is invalid")
	ErrTokenExpired = errors.New("token is expired")
	ErrLoginYet     = errors.New("have not logged in yet")
)

type AuthnHandler struct {
	ignorePaths map[string]struct{}
	cmd         redis.Cmdable
	jwt         jwt.Handler
}

func NewAuthnHandler(cmd redis.Cmdable, jwt jwt.Handler) *AuthnHandler {
	return &AuthnHandler{
		ignorePaths: make(map[string]struct{}),
		cmd:         cmd,
		jwt:         jwt,
	}
}

func (a *AuthnHandler) IgnorePaths(path string) *AuthnHandler {
	a.ignorePaths[path] = struct{}{}
	return a
}

func (a *AuthnHandler) Authn() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := a.ignorePaths[c.Request.URL.Path]; ok {
			c.Next()
			return
		}

		token := a.jwt.ExtractToken(c)

		claims, err := ParseToken(token)
		if err != nil {
			errCode := handleTokenError(err)
			response.Error(c, errCode)
			c.Abort()
			return
		}

		if claims.UserAgent != c.Request.UserAgent() {
			response.Error(c, er.NewBizError(constant.ErrUserUnauthorized))
			c.Abort()
			return
		}

		err = a.jwt.CheckSession(c, claims.SSId)
		if err != nil {
			response.Error(c, er.NewBizError(constant.ErrUserSessExpired))
			c.Abort()
			return
		}

		c.Set("claims", claims)

		c.Next()
	}
}

func ParseToken(tk string) (*jwt.Claims, error) {
	tokenClaims, err := sjwt.ParseWithClaims(tk, &jwt.Claims{}, func(tk *sjwt.Token) (interface{}, error) {
		return jwt.SecretKey, nil
	})
	if err != nil {
		var ve *sjwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&sjwt.ValidationErrorMalformed != 0 {
				return nil, ErrTokenInvalid
			} else if ve.Errors&(sjwt.ValidationErrorExpired) != 0 {
				return nil, ErrTokenExpired
			} else if ve.Errors&(sjwt.ValidationErrorNotValidYet) != 0 {
				return nil, ErrLoginYet
			}
		}
		return nil, err
	}
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*jwt.Claims); ok && tokenClaims.Valid {
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
