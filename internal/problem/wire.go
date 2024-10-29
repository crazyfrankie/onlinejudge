//go:build wireinject

package problem

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"oj/internal/problem/repository"
	"oj/internal/problem/repository/cache"
	"oj/internal/problem/repository/dao"
	"oj/internal/problem/service"
	"oj/internal/problem/web"
)

func InitProblemRepo(cmd redis.Cmdable, db *gorm.DB) repository.ProblemRepository {
	wire.Build(
		cache.NewProblemCache,
		dao.NewProblemDao,

		repository.NewProblemRepository,
	)
	return new(repository.CacheProblemRepo)
}

func InitProblemHandler(cmd redis.Cmdable, db *gorm.DB) *web.ProblemHandler {
	wire.Build(
		InitProblemRepo,

		service.NewProblemService,

		web.NewProblemHandler,
	)
	return new(web.ProblemHandler)
}
