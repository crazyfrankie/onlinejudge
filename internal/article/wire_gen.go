// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package article

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
	"oj/internal/article/repository"
	"oj/internal/article/repository/dao"
	"oj/internal/article/service"
	"oj/internal/article/web"
)

// Injectors from wire.go:

func InitArticleHandler(db *gorm.DB) *web.ArticleHandler {
	articleDao := dao.NewArticleDao(db)
	articleRepository := repository.NewArticleRepository(articleDao)
	articleService := service.NewArticleService(articleRepository)
	logger := InitLog()
	articleHandler := web.NewArticleHandler(articleService, logger)
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
