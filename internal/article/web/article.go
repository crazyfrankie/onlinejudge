package web

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	"github.com/crazyfrankie/onlinejudge/common/constant"
	"github.com/crazyfrankie/onlinejudge/common/errors"
	"github.com/crazyfrankie/onlinejudge/common/response"
	"github.com/crazyfrankie/onlinejudge/internal/article/domain"
	"github.com/crazyfrankie/onlinejudge/internal/article/service"
	"github.com/crazyfrankie/onlinejudge/internal/middleware/jwt"
)

const (
	Biz = "article"
)

type ArticleHandler struct {
	svc      service.ArticleService
	interSvc *service.InteractiveService
}

func NewArticleHandler(svc service.ArticleService, interSvc *service.InteractiveService) *ArticleHandler {
	return &ArticleHandler{
		svc:      svc,
		interSvc: interSvc,
	}
}

func (ctl *ArticleHandler) RegisterRoute(r *gin.Engine) {
	artGroup := r.Group("api/articles")
	{
		artGroup.POST("pub/list", ctl.PubList())
		artGroup.POST("pub/detail/:id", ctl.PubDetail())
		artGroup.POST("like", ctl.Like())
	}
}

func (ctl *ArticleHandler) PubList() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ListReq
		if err := c.ShouldBind(&req); err != nil {
			response.Error(c, errors.NewBizError(constant.ErrArticleInvalidParams))
			return
		}

		res, err := ctl.svc.PubList(c.Request.Context(), req.Offset, req.Limit)
		if err != nil {
			response.Error(c, err)
			return
		}

		var resp []PubListResp
		for _, art := range res {
			resp = append(resp, PubListResp{
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

func (ctl *ArticleHandler) PubDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := c.MustGet("claims")
		claim := claims.(*jwt.Claims)

		artID := c.Param("id")

		var err error
		var eg errgroup.Group
		var art domain.Article
		eg.Go(func() error {
			art, err = ctl.svc.PubDetail(c.Request.Context(), claim.Id, artID)
			return err
		})

		var inter domain.Interactive
		eg.Go(func() error {
			inter, err = ctl.interSvc.GetInteractive(c.Request.Context(), Biz, artID, claim.Id)
			return err
		})

		err = eg.Wait()
		if err != nil {
			response.Error(c, err)
			return
		}

		interResp := Interactive{
			LikeCnt: inter.LikeCnt,
			ReadCnt: inter.ReadCnt,
		}
		resp := PubDetailResp{
			ID:         art.ID,
			Title:      art.Title,
			Content:    art.Content,
			AuthorID:   art.Author.Id,
			AuthorName: art.Author.Name,
			Ctime:      art.Ctime.String(),
			Utime:      art.Utime.String(),
			Status:     art.Status.ToUint8(),
			Inter:      interResp,
			Liked:      inter.Liked,
		}

		response.Success(c, resp)
	}
}

func (ctl *ArticleHandler) Like() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LikeReq
		if err := c.ShouldBind(&req); err != nil {
			response.Error(c, errors.NewBizError(constant.ErrArticleInvalidParams))
			return
		}

		claims := c.MustGet("claims")
		claim := claims.(*jwt.Claims)

		var err error
		if req.Like {
			err = ctl.interSvc.Like(c.Request.Context(), Biz, req.ID, claim.Id)
		} else {
			err = ctl.interSvc.CancelLike(c.Request.Context(), Biz, req.ID, claim.Id)
		}

		if err != nil {
			response.Error(c, err)
		}

		response.Success(c, nil)
	}
}
