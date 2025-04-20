package web

import (
	"github.com/crazyfrankie/onlinejudge/pkg/zapx"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	"github.com/crazyfrankie/onlinejudge/common/constant"
	"github.com/crazyfrankie/onlinejudge/common/errors"
	"github.com/crazyfrankie/onlinejudge/common/response"
	"github.com/crazyfrankie/onlinejudge/internal/article/domain"
	"github.com/crazyfrankie/onlinejudge/internal/article/service"
	"github.com/crazyfrankie/onlinejudge/internal/auth"
)

const (
	Biz = "article"
)

type ArticleHandler struct {
	svc      service.ArticleService
	interSvc *service.InteractiveService
	logger   *zapx.Logger
}

func NewArticleHandler(svc service.ArticleService, interSvc *service.InteractiveService, logger *zapx.Logger) *ArticleHandler {
	return &ArticleHandler{
		svc:      svc,
		interSvc: interSvc,
		logger:   logger,
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
		name := "onlinejudge/Article/PubList"
		var req ListReq
		if err := c.ShouldBind(&req); err != nil {
			response.ErrorWithLog(c, name, bizError, errors.NewBizError(constant.ErrArticleInvalidParams))
			return
		}

		res, err := ctl.svc.PubList(c.Request.Context(), req.Offset, req.Limit)
		if err != nil {
			response.ErrorWithLog(c, name, bizError, err)
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

		response.SuccessWithLog(c, resp, name, success)
	}
}

func (ctl *ArticleHandler) PubDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := "onlinejudge/Article/PubDetail"

		claims := c.MustGet("claims")
		claim := claims.(*auth.Claims)

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
			response.ErrorWithLog(c, name, bizError, err)
			return
		}

		interResp := Interactive{
			LikeCnt: inter.LikeCnt,
			ReadCnt: inter.ReadCnt + 1,
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

		response.SuccessWithLog(c, resp, name, success)
	}
}

func (ctl *ArticleHandler) Like() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := "onlinejudge/Article/Like"
		var req LikeReq
		if err := c.ShouldBind(&req); err != nil {
			response.ErrorWithLog(c, name, bizError, errors.NewBizError(constant.ErrArticleInvalidParams))
			return
		}

		claims := c.MustGet("claims")
		claim := claims.(*auth.Claims)

		var err error
		if req.Like {
			err = ctl.interSvc.Like(c.Request.Context(), Biz, req.ID, claim.Id)
		} else {
			err = ctl.interSvc.CancelLike(c.Request.Context(), Biz, req.ID, claim.Id)
		}

		if err != nil {
			response.ErrorWithLog(c, name, bizError, err)
		}

		response.SuccessWithLog(c, nil, name, success)
	}
}
