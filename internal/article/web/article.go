package web

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"oj/internal/article/domain"
	"oj/internal/article/service"
	ijwt "oj/internal/user/middleware/jwt"
)

type ArticleHandler struct {
	svc service.ArticleService
	l   *zap.Logger
}

func NewArticleHandler(svc service.ArticleService, l *zap.Logger) *ArticleHandler {
	return &ArticleHandler{
		svc: svc,
		l:   l,
	}
}

func (ctl *ArticleHandler) RegisterRoute(r *gin.Engine) {
	postGroup := r.Group("articles")
	{
		postGroup.POST("/edit", ctl.Edit())
		postGroup.POST("/publish", ctl.Publish())
	}
}

func (ctl *ArticleHandler) Edit() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ArticleReq
		if err := c.Bind(&req); err != nil {
			return
		}

		claims, ok := c.Get("claims")
		if !ok {
			c.JSON(http.StatusInternalServerError, Result[uint64]{
				Data: 0,
				Msg:  "system error:未发现作者的用户信息",
			})
			return
		}
		claim := claims.(ijwt.Claims)

		id, err := ctl.svc.Save(c.Request.Context(), req.toDomain(claim.Id))
		if err != nil {
			ctl.l.Error("创建/更新帖子:系统错误")
			c.JSON(http.StatusInternalServerError, Result[uint64]{
				Data: 0,
				Msg:  "system error",
			})
			return
		}

		ctl.l.Info("帖子创建/更新成功")
		c.JSON(http.StatusOK, Result[uint64]{
			Data: id,
			Msg:  "OK",
		})
	}
}

func (ctl *ArticleHandler) Publish() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ArticleReq
		if err := c.Bind(&req); err != nil {
			return
		}

		claims, ok := c.Get("claims")
		if !ok {
			c.JSON(http.StatusInternalServerError, Result[uint64]{
				Data: 0,
				Msg:  "system error:未发现作者的用户信息",
			})
			return
		}
		claim := claims.(ijwt.Claims)

		id, err := ctl.svc.Publish(c.Request.Context(), req.toDomain(claim.Id))
		if err != nil {
			ctl.l.Error("发布帖子:系统错误")
			c.JSON(http.StatusInternalServerError, Result[uint64]{
				Data: 0,
				Msg:  "system error",
			})
			return
		}

		ctl.l.Info("帖子发布成功")
		c.JSON(http.StatusOK, Result[uint64]{
			Data: id,
			Msg:  "OK",
		})
	}
}

type ArticleReq struct {
	ID      uint64 `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (req ArticleReq) toDomain(uid uint64) domain.Article {
	return domain.Article{
		ID: req.ID,
		Author: domain.Author{
			Id: uid,
		},
		Title:   req.Title,
		Content: req.Content,
	}
}
