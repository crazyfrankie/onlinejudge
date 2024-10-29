//go:build wireinject

package integration

import (
	"github.com/google/wire"
	"oj/internal/article/service"
)

func InitArticleService() *service.ArticleService {
	wire.Build(
		service.NewArticleService,
	)
	return new(service.ArticleService)
}
