package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/onlinejudge/common/constant"
	er "github.com/crazyfrankie/onlinejudge/common/errors"
	"github.com/crazyfrankie/onlinejudge/common/response"
	"github.com/crazyfrankie/onlinejudge/internal/middleware/jwt"
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

		claims, err := a.jwt.ParseToken(token)
		if err != nil {
			errCode := a.jwt.HandleTokenError(err)
			response.Error(c, errCode)
			c.Abort()
			return
		}

		if claims.UserAgent != c.Request.UserAgent() {
			response.Error(c, er.NewBizError(constant.ErrUserUnauthorized))
			c.Abort()
			return
		}

		if err = a.jwt.CheckSession(c, claims.Id, claims.SSId); err != nil {
			response.Error(c, er.NewBizError(constant.ErrUserSessExpired))
			c.Abort()
			return
		}

		if err = a.jwt.TryRefresh(c); err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}

		c.Set("claims", claims)

		c.Next()
	}
}
