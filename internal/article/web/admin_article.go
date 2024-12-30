package web

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	"github.com/crazyfrankie/onlinejudge/common/constant"
	"github.com/crazyfrankie/onlinejudge/common/errors"
	"github.com/crazyfrankie/onlinejudge/common/response"
	"github.com/crazyfrankie/onlinejudge/internal/article/domain"
	"github.com/crazyfrankie/onlinejudge/internal/article/service"
	"github.com/crazyfrankie/onlinejudge/internal/middleware/jwt"
)

type AdminHandler struct {
	svc service.ArticleService
}

func NewAdminHandler(svc service.ArticleService) *AdminHandler {
	return &AdminHandler{
		svc: svc,
	}
}

func (ctl *AdminHandler) RegisterRoute(r *gin.Engine) {
	admin := r.Group("api/articles")
	{
		admin.POST("save", ctl.Edit())
		admin.POST("list", ctl.List())
		admin.POST("detail/:id", ctl.Detail())
		admin.POST("publish", ctl.Publish())
	}
}

func (ctl *AdminHandler) Edit() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ArticleReq
		if err := c.ShouldBind(&req); err != nil {
			response.Error(c, errors.NewBizError(constant.ErrArticleInvalidParams))
			return
		}

		claims := c.MustGet("claims")
		claim := claims.(jwt.Claims)

		id, err := ctl.svc.SaveDraft(c.Request.Context(), req.toDomain(claim.Id))
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, id)
	}
}

func (ctl *AdminHandler) Publish() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ArticleReq
		if err := c.ShouldBind(&req); err != nil {
			response.Error(c, errors.NewBizError(constant.ErrArticleInvalidParams))
			return
		}

		claims := c.MustGet("claims")
		claim := claims.(jwt.Claims)

		id, err := ctl.svc.Publish(c.Request.Context(), req.toDomain(claim.Id))
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, id)
	}
}

func (ctl *AdminHandler) WithDraw() gin.HandlerFunc {
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
			response.Error(c, err)
			return
		}

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

func (ctl *AdminHandler) List() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ListReq
		if err := c.ShouldBind(&req); err != nil {
			response.Error(c, errors.NewBizError(constant.ErrArticleInvalidParams))
			return
		}

		res, err := ctl.svc.List(c.Request.Context(), req.Offset, req.Limit)
		if err != nil {
			response.Error(c, err)
			return
		}

		var resp []ListResp
		for _, art := range res {
			resp = append(resp, ListResp{
				ID:     art.ID,
				Title:  art.Title,
				Status: art.Status.ToUint8(),
				Ctime:  art.Ctime.String(),
				Utime:  art.Utime.String(),
			})
		}

		response.Success(c, resp)
	}
}

func (ctl *AdminHandler) Detail() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := c.MustGet("claims")
		claim := claims.(*jwt.Claims)

		artID := c.Param("id")

		var err error
		var eg errgroup.Group
		var art domain.Article
		eg.Go(func() error {
			art, err = ctl.svc.Detail(c.Request.Context(), claim.Id, artID)
			return err
		})

		err = eg.Wait()
		if err != nil {
			response.Error(c, err)
			return
		}

		resp := DetailResp{
			ID:      art.ID,
			Title:   art.Title,
			Content: art.Content,
			Ctime:   art.Ctime.String(),
			Utime:   art.Utime.String(),
			Status:  art.Status.ToUint8(),
		}

		response.Success(c, resp)
	}
}
