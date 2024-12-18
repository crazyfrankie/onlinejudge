//go:build wireinject

package article

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"oj/internal/article/repository"
	"oj/internal/article/repository/cache"
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

func InitArticleHandler(db *gorm.DB, cmd redis.Cmdable) *web.ArticleHandler {
	wire.Build(
		dao.NewArticleDao,
		dao.NewInteractiveDao,
		cache.NewInteractiveCache,

		repository.NewArticleRepository,
		repository.NewInteractiveArtRepository,

		service.NewArticleService,
		service.NewInteractiveService,

		InitLog,

		web.NewArticleHandler,
	)
	return new(web.ArticleHandler)
}
