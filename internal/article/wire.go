//go:build wireinject

package article

import (
	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"oj/internal/article/repository"
	"oj/internal/article/repository/dao"
	"oj/internal/article/service"
	"oj/internal/article/web"
)

func InitLog() *zap.Logger {
	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return log
}

func InitArticleHandler(db *gorm.DB) *web.ArticleHandler {
	wire.Build(
		dao.NewArticleDao,

		repository.NewArticleRepository,

		service.NewArticleService,

		InitLog,

		web.NewArticleHandler,
	)
	return new(web.ArticleHandler)
}
