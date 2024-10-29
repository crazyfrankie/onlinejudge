//go:build wireinject

package article

import (
	"github.com/google/wire"
	"oj/internal/article/service"
	"oj/internal/article/web"
)

func InitArticleHandler() *web.ArticleHandler {
	wire.Build(
		service.NewArticleService,
		web.NewArticleHandler,
	)
	return new(web.ArticleHandler)
}
