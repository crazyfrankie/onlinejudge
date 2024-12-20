// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package article

import (
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"oj/internal/article/repository"
	"oj/internal/article/repository/cache"
	"oj/internal/article/repository/dao"
	"oj/internal/article/service"
	"oj/internal/article/web"
)

// Injectors from wire.go:

func InitArticleHandler(db *gorm.DB, cmd redis.Cmdable) *web.ArticleHandler {
	gormArticleDao := dao.NewArticleDao(db)
	articleRepository := repository.NewArticleRepository(gormArticleDao)
	logger := InitLog()
	articleService := service.NewArticleService(articleRepository, logger)
	interactiveDao := dao.NewInteractiveDao(db)
	interactiveCache := cache.NewInteractiveCache(cmd)
	interactiveArtRepository := repository.NewInteractiveArtRepository(interactiveDao, interactiveCache)
	interactiveService := service.NewInteractiveService(interactiveArtRepository)
	articleHandler := web.NewArticleHandler(articleService, logger, interactiveService)
	return articleHandler
}

// wire.go:

func InitLog() *zap.Logger {
	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return log
}
