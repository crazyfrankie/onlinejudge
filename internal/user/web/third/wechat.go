package third

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"oj/internal/user/domain"
	ijwt "oj/internal/user/middleware/jwt"
	"oj/internal/user/service"
	"oj/internal/user/service/oauth/wechat"
	"oj/internal/user/web"
)

type OAuthWeChatHandler struct {
	svc     wechat.Service
	userSvc service.UserService
	ijwt.Handler
	stateKey []byte
}

type StateClaims struct {
	State string
	jwt.StandardClaims
}

func NewOAuthHandler(svc wechat.Service, jwtHdl ijwt.Handler, userSvc service.UserService) *OAuthWeChatHandler {
	return &OAuthWeChatHandler{
		svc:      svc,
		Handler:  jwtHdl,
		userSvc:  userSvc,
		stateKey: []byte("KsS2X1CgFT4bi3BRRIxLk5jjiUBj8wxF"),
	}
}

func (h *OAuthWeChatHandler) RegisterRoute(r *server.Hertz) {
	oauthGroup := r.Group("/oauth/wechat")
	{
		oauthGroup.GET("/url", h.AuthUrl())
		oauthGroup.Any("/callback", h.CallBack())
	}
}

func (h *OAuthWeChatHandler) AuthUrl() app.HandlerFunc {
	return func(c *app.RequestContext) {
		state := uuid.New().String()
		url, err := h.svc.AuthURL(c.Request.Context(), state)
		if err != nil {
			c.JSON(http.StatusInternalServerError, web.GetResponse(web.WithStatus(http.StatusInternalServerError), web.WithMsg("get url failed")))
			return
		}

		if err := h.SetCookie(c, state); err != nil {
			c.JSON(http.StatusInternalServerError, web.GetResponse(web.WithStatus(http.StatusInternalServerError), web.WithMsg("system error")))
			return
		}

		c.JSON(http.StatusOK, web.GetResponse(web.WithStatus(http.StatusOK), web.WithData(url)))
	}
}

func (h *OAuthWeChatHandler) SetCookie(c *app.RequestContext, state string) error {
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

func (h *OAuthWeChatHandler) CallBack() app.HandlerFunc {
	return func(c *app.RequestContext) {
		code := c.Query("code")

		err := h.VerifyState(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, web.GetResponse(web.WithStatus(http.StatusInternalServerError), web.WithMsg("system error")))
			return
		}

		info, err := h.svc.VerifyCode(c.Request.Context(), code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, web.GetResponse(web.WithStatus(http.StatusInternalServerError), web.WithMsg("system error")))
			return
		}

		user, err := h.userSvc.FindOrCreateByWechat(c.Request.Context(), domain.WeChatInfo{
			OpenID:  info.OpenID,
			UnionID: info.UnionID,
		})

		err = h.Handler.SetLoginToken(c, 0, user.Id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, web.GetResponse(web.WithStatus(http.StatusInternalServerError), web.WithMsg("system error")))
			return
		}

		c.JSON(http.StatusOK, web.GetResponse(web.WithStatus(http.StatusOK), web.WithMsg("login successfully")))
	}
}

func (h *OAuthWeChatHandler) VerifyState(c *app.RequestContext) error {
	state := c.Query("state")
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
