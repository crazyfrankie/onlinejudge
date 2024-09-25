package web

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"oj/internal/middleware"
	"oj/internal/user/domain"
	"oj/internal/user/service"
	"oj/internal/user/service/oauth/wechat"
)

type OAuthWeChatHandler struct {
	svc      wechat.Service
	userSvc  service.UserService
	tokenGen middleware.TokenGenerator
}

func NewOAuthHandler(svc wechat.Service, tokenGen middleware.TokenGenerator, userSvc service.UserService) *OAuthWeChatHandler {
	return &OAuthWeChatHandler{
		svc:      svc,
		tokenGen: tokenGen,
		userSvc:  userSvc,
	}
}

func (h *OAuthWeChatHandler) RegisterRoute(r *gin.Engine) {
	oauthGroup := r.Group("/oauth/wechat")
	{
		oauthGroup.GET("/authurl", h.AuthUrl())
		oauthGroup.Any("/callback", h.CallBack())
	}
}

func (h *OAuthWeChatHandler) AuthUrl() gin.HandlerFunc {
	return func(c *gin.Context) {
		url, err := h.svc.AuthURL(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusBadRequest, "get url failed")
			return
		}
		c.JSON(http.StatusOK, url)
	}
}

func (h *OAuthWeChatHandler) CallBack() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")
		info, err := h.svc.VerifyCode(c.Request.Context(), code, state)
		if err != nil {
			c.JSON(http.StatusBadRequest, "system error")
			return
		}

		user, err := h.userSvc.FindOrCreateByWechat(c.Request.Context(), domain.WeChatInfo{
			OpenID:  info.OpenID,
			UnionID: info.UnionID,
		})

		userAgent := c.GetHeader("User-Agent")

		token, err := h.tokenGen.GenerateToken(0, user.Id, userAgent)
		if err != nil {
			c.JSON(http.StatusBadRequest, "system error")
			return
		}

		c.JSON(http.StatusOK, token)
	}
}
