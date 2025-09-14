package mws

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/onlinejudge/common/constant"
	"github.com/crazyfrankie/onlinejudge/common/errors"
	"github.com/crazyfrankie/onlinejudge/common/response"
	"github.com/crazyfrankie/onlinejudge/infra/contract/token"
)

type Authorizer interface {
	Authorize(subject, object, action string) (bool, error)
}

type AuthzHandler struct {
	auth Authorizer
}

func NewAuthzHandler(auth Authorizer) *AuthzHandler {
	return &AuthzHandler{auth: auth}
}

func (a *AuthzHandler) Authz() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := c.Get("claims")
		if !ok {
			c.Next()
			return
		}
		claim := claims.(*token.Claims)
		sub := strconv.FormatUint(claim.Id, 10)
		obj := c.FullPath()
		act := "CALL"

		if allowed, err := a.auth.Authorize(sub, obj, act); err != nil || !allowed {
			response.Error(c, errors.NewBizError(constant.ErrUserForbidden))
			c.Abort()
			return
		}

		c.Next()
	}
}
