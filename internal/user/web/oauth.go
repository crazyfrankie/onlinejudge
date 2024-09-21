package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"oj/internal/middleware"
	"oj/internal/user/service/oauth/wechat"
)

type OAuthWeChatHandler struct {
	svc      wechat.Service
	tokenGen middleware.TokenGenerator
}

func NewOAuthHandler(svc wechat.Service, tokenGen middleware.TokenGenerator) *OAuthWeChatHandler {
	return &OAuthWeChatHandler{
		svc:      svc,
		tokenGen: tokenGen,
	}
}

func (h *OAuthWeChatHandler) RegisterRoute(r *gin.Engine) {
	oauthGroup := r.Group("/oauth/wechat")
	{
		oauthGroup.GET("/authurl", h.AuthUrl())
		//oauthGroup.Any("/callback", h.CallBack())
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

//func (h *OAuthWeChatHandler) CallBack() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		code := c.Query("code")
//		state := c.Query("state")
//		info, err := h.svc.VerifyCode(c.Request.Context(), code, state)
//		if err != nil {
//			c.JSON(http.StatusBadRequest, "system error")
//			return
//		}
//
//		h.tokenGen.GenerateToken()
//	}
//}
