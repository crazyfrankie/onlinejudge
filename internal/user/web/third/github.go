package third

import (
	"errors"
	"fmt"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"

	"github.com/crazyfrankie/onlinejudge/common/constant"
	er "github.com/crazyfrankie/onlinejudge/common/errors"
	"github.com/crazyfrankie/onlinejudge/common/response"
	"github.com/crazyfrankie/onlinejudge/internal/auth"
	"github.com/crazyfrankie/onlinejudge/internal/user/domain"
	"github.com/crazyfrankie/onlinejudge/internal/user/service"
	"github.com/crazyfrankie/onlinejudge/internal/user/service/oauth/github"
)

const (
	bizError = "biz error"
	success  = "success handle"
)

type OAuthGithubHandler struct {
	svc      github.Service
	userSvc  service.UserService
	jwt      auth.JWTHandler
	stateKey []byte
}

func NewOAuthGithubHandler(svc github.Service, userSvc service.UserService, jwtHdl auth.JWTHandler) *OAuthGithubHandler {
	return &OAuthGithubHandler{
		svc:      svc,
		userSvc:  userSvc,
		jwt:      jwtHdl,
		stateKey: []byte("KsS2X1CgFT4bi3BRRIxLk5jjiUBj8wxF"),
	}
}

func (h *OAuthGithubHandler) RegisterRoute(r *gin.Engine) {
	oauthGroup := r.Group("/oauth/github")
	{
		oauthGroup.GET("/url", h.GitAuthUrl())
		oauthGroup.Any("/callback", h.CallBack())
	}
}

func (h *OAuthGithubHandler) GitAuthUrl() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := "onlinejudge/OAuth2/Github/GitAuthUrl"
		state := uuid.New().String()

		url, err := h.svc.AuthUrl(c.Request.Context(), state)
		if err != nil {
			response.ErrorWithLog(c, name, bizError, er.NewBizError(constant.ErrUserInternalServer))
			return
		}

		if err := h.SetCookie(c, state); err != nil {
			response.ErrorWithLog(c, name, bizError, er.NewBizError(constant.ErrUserInternalServer))
			return
		}

		response.SuccessWithLog(c, url, name, success)
	}
}

func (h *OAuthGithubHandler) SetCookie(c *gin.Context, state string) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, StateClaims{
		State: state,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 10).Unix(),
		},
	})
	tokenStr, err := token.SignedString(h.stateKey)
	if err != nil {
		return err
	}
	c.SetCookie("jwt-state", tokenStr, 600, "/", "", false, true)
	return nil
}

func (h *OAuthGithubHandler) CallBack() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := "onlinejudge/OAuth2/Github/CallBack"
		code := c.Query("code")

		err := h.VerifyState(c)
		if err != nil {
			response.ErrorWithLog(c, name, bizError, er.NewBizError(constant.ErrInternalServer))
			return
		}

		var res github.Result
		res, err = h.svc.VerifyCode(c.Request.Context(), code)
		if err != nil {
			response.ErrorWithLog(c, name, bizError, er.NewBizError(constant.ErrUserInternalServer))
			return
		}

		var info domain.GithubInfo
		info, err = h.svc.AcquireUserInfo(c.Request.Context(), res.AccessToken)
		if err != nil {
			response.ErrorWithLog(c, name, bizError, er.NewBizError(constant.ErrUserInternalServer))
			return
		}

		user, err := h.userSvc.FindOrCreateByGithub(c.Request.Context(), info.Id)

		err = h.jwt.SetLoginToken(c, user.Id)
		if err != nil {
			response.ErrorWithLog(c, name, bizError, er.NewBizError(constant.ErrUserInternalServer))
			return
		}

		url := "http://localhost:8081"
		c.Redirect(http.StatusFound, url)
	}
}

func (h *OAuthGithubHandler) VerifyState(c *gin.Context) error {
	state := c.Query("jwt-state")
	jwtState, err := c.Cookie("jwt-state")
	if err != nil {
		return fmt.Errorf("拿不到 state 的 cookie, %w", err)
	}

	var sc StateClaims
	token, err := jwt.ParseWithClaims(jwtState, &sc, func(token *jwt.Token) (interface{}, error) {
		return h.stateKey, nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("token 已经过期, %w", err)
	}

	if sc.State != state {
		return errors.New("state 不相等")
	}

	return nil
}
