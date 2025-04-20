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
	"github.com/crazyfrankie/onlinejudge/internal/user/service/oauth/wechat"
)

type OAuthWeChatHandler struct {
	svc      wechat.Service
	userSvc  service.UserService
	jwt      auth.JWTHandler
	stateKey []byte
}

type StateClaims struct {
	State string
	jwt.StandardClaims
}

func NewOAuthWeChatHandler(svc wechat.Service, jwtHdl auth.JWTHandler, userSvc service.UserService) *OAuthWeChatHandler {
	return &OAuthWeChatHandler{
		svc:      svc,
		jwt:      jwtHdl,
		userSvc:  userSvc,
		stateKey: []byte("KsS2X1CgFT4bi3BRRIxLk5jjiUBj8wxF"),
	}
}

func (h *OAuthWeChatHandler) RegisterRoute(r *gin.Engine) {
	oauthGroup := r.Group("/oauth/wechat")
	{
		oauthGroup.GET("/url", h.AuthUrl())
		oauthGroup.Any("/callback", h.CallBack())
	}
}

func (h *OAuthWeChatHandler) AuthUrl() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := "onlinejudge/OAuth2/Wechat/AuthUrl"
		state := uuid.New().String()
		url, err := h.svc.AuthURL(c.Request.Context(), state)
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

func (h *OAuthWeChatHandler) SetCookie(c *gin.Context, state string) error {
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
	c.SetCookie("jwt-state", tokenStr, 600, "/oauth/wechat/callback", "", false, true)
	return nil
}

func (h *OAuthWeChatHandler) CallBack() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := "onlinejudge/OAuth2/Wechat/CallBack"
		code := c.Query("code")

		err := h.VerifyState(c)
		if err != nil {
			response.ErrorWithLog(c, name, bizError, er.NewBizError(constant.ErrUserInternalServer))
			return
		}

		info, err := h.svc.VerifyCode(c.Request.Context(), code)
		if err != nil {
			response.ErrorWithLog(c, name, bizError, er.NewBizError(constant.ErrUserInternalServer))
			return
		}

		user, err := h.userSvc.FindOrCreateByWechat(c.Request.Context(), domain.WeChatInfo{
			OpenID:  info.OpenID,
			UnionID: info.UnionID,
		})

		err = h.jwt.SetLoginToken(c, user.Id)
		if err != nil {
			response.ErrorWithLog(c, name, bizError, er.NewBizError(constant.ErrUserInternalServer))
			return
		}

		url := "http://localhost:8081"
		c.Redirect(http.StatusFound, url)
	}
}

func (h *OAuthWeChatHandler) VerifyState(c *gin.Context) error {
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
