package web

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
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

func (ctl *ArticleHandler) RegisterRoute(r *server.Hertz) {
	postGroup := r.Group("articles")
	{
		postGroup.POST("/edit", ctl.Edit())
		postGroup.POST("/publish", ctl.Publish())
	}
}

func (ctl *ArticleHandler) Edit() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
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

		id, err := ctl.svc.SaveDraft(ctx, req.toDomain(claim.Id))
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

func (ctl *ArticleHandler) Publish() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
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

		id, err := ctl.svc.Publish(ctx, req.toDomain(claim.Id))
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

// 首先库（表）分制作库（表）和线上库（表）
// 对于 web 层来说
// Edit:作者新建帖子/修改已有帖子并保存到制作库
// Publish:作者新建帖子/修改已有帖子并保存到线上库
// 到 service 层
// SaveDraft:如果存在文章 Id:代表是修改并保存,如果不存在是新建并保存
// Publish:如果存在文章 Id:代表是修改并发布,如果不存在是新建并发布
// 到 Repo 层
// 由于有制作库和线上库,实际就是查询时查哪个库,将 Repo 分为 AuthorRepo 和 ReaderRepo
// 在 service 层面做数据同步
// Save 接口啥也不用改，因为它只改制作库
// Publish 接口先改制作库，然后同步到线上库
