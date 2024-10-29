package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"oj/internal/article/domain"
	"oj/internal/article/service"
)

type ArticleHandler struct {
	svc *service.ArticleService
}

func NewArticleHandler(svc *service.ArticleService) *ArticleHandler {
	return &ArticleHandler{
		svc: svc,
	}
}

func (ctl *ArticleHandler) RegisterRoute(r *gin.Engine) {
	postGroup := r.Group("posts")
	{
		postGroup.POST("/edit", ctl.Edit())
	}
}

func (ctl *ArticleHandler) Edit() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		}
		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}

		id, err := ctl.svc.Save(c.Request.Context(), domain.Article{
			Title:   req.Title,
			Content: req.Content,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, Result[uint64]{
				Data: 0,
				Msg:  "system error",
			})
			return
		}

		c.JSON(http.StatusOK, Result[uint64]{
			Data: id,
			Msg:  "OK",
		})
	}
}
