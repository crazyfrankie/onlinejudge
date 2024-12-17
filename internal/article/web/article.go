package web

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"oj/common/constant"
	"oj/common/errors"
	"oj/common/response"
	"oj/internal/article/domain"
	"oj/internal/article/service"
	"oj/internal/user/middleware/jwt"
)

type ArticleHandler struct {
	svc      service.ArticleService
	interSvc *service.InteractiveService
	l        *zap.Logger
	biz      string
}

func NewArticleHandler(svc service.ArticleService, l *zap.Logger, interSvc *service.InteractiveService) *ArticleHandler {
	return &ArticleHandler{
		svc:      svc,
		interSvc: interSvc,
		l:        l,
		biz:      "article",
	}
}

func (ctl *ArticleHandler) RegisterRoute(r *gin.Engine) {
	postGroup := r.Group("articles")
	{
		postGroup.POST("/edit", ctl.Edit())
		postGroup.POST("/publish", ctl.Publish())
		postGroup.POST("/:id/withdraw", ctl.WithDraw())
		postGroup.POST("/:id/detail", ctl.Detail())
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
			response.Error(c, errors.NewBusinessError(constant.ErrInternalServer))
			return
		}
		claim := claims.(jwt.Claims)

		id, err := ctl.svc.SaveDraft(c.Request.Context(), req.toDomain(claim.Id))
		if err != nil {
			ctl.l.Error("创建/更新帖子:系统错误")
			response.Error(c, err)
			return
		}

		ctl.l.Info("帖子创建/更新成功")
		response.Success(c, id)
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
			response.Error(c, errors.NewBusinessError(constant.ErrInternalServer))
			return
		}
		claim := claims.(jwt.Claims)

		id, err := ctl.svc.Publish(c.Request.Context(), req.toDomain(claim.Id))
		if err != nil {
			ctl.l.Error("发布帖子:系统错误")
			response.Error(c, err)
			return
		}

		ctl.l.Info("帖子发布成功")
		response.Success(c, id)
	}
}

func (ctl *ArticleHandler) WithDraw() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Query("id")

		claims := c.MustGet("claims")
		claim := claims.(jwt.Claims)

		Id, _ := strconv.Atoi(id)
		err := ctl.svc.WithDraw(c.Request.Context(), domain.Article{
			ID: uint64(Id),
			Author: domain.Author{
				Id: claim.Id,
			},
		})
		if err != nil {
			ctl.l.Error("撤回帖子:系统错误")
			response.Error(c, err)
			return
		}

		ctl.l.Info("帖子撤回成功")
		response.Success(c, nil)
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

func (ctl *ArticleHandler) List() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ListReq
		if err := c.Bind(&req); err != nil {
			return
		}

		claims := c.MustGet("claims")
		claim := claims.(*jwt.Claims)

		res, err := ctl.svc.List(c.Request.Context(), claim.Id, req.Offset, req.Limit)
		if err != nil {
			response.Error(c, err)
			return
		}

		var resp []ListResp
		for _, art := range res {
			resp = append(resp, ListResp{
				ID:         art.ID,
				Title:      art.Title,
				Abstract:   art.Abstract(),
				AuthorID:   art.Author.Id,
				AuthorName: art.Author.Name,
				Status:     art.Status.ToUint8(),
				Ctime:      art.Ctime.String(),
				Utime:      art.Utime.String(),
			})
		}

		response.Success(c, resp)
	}
}

func (ctl *ArticleHandler) Detail() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := c.MustGet("claims")
		claim := claims.(*jwt.Claims)

		artID := c.Param("id")

		art, err := ctl.svc.Detail(c.Request.Context(), claim.Id, artID)
		if err != nil {
			response.Error(c, err)
		}

		// 增加阅读计数
		go func() {
			er := ctl.interSvc.IncrReadCnt(c.Request.Context(), ctl.biz, artID)
			if er != nil {
				log.Printf("增加阅读计数失败:aid:%s", artID)
			}
		}()

		resp := DetailResp{
			ID:         art.ID,
			Title:      art.Title,
			Content:    art.Content,
			AuthorID:   art.Author.Id,
			AuthorName: art.Author.Name,
			Ctime:      art.Ctime.String(),
			Utime:      art.Utime.String(),
			Status:     art.Status.ToUint8(),
		}
		response.Success(c, resp)
	}
}
