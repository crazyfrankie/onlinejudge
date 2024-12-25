//go:build wireinject

package problem

import (
	"github.com/crazyfrankie/onlinejudge/internal/problem/repository"
	"github.com/crazyfrankie/onlinejudge/internal/problem/repository/cache"
	"github.com/crazyfrankie/onlinejudge/internal/problem/repository/dao"
	"github.com/crazyfrankie/onlinejudge/internal/problem/service"
	"github.com/crazyfrankie/onlinejudge/internal/problem/web"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func InitModule(cmd redis.Cmdable, db *gorm.DB) *Module {
	wire.Build(
		cache.NewProblemCache,
		dao.NewProblemDao,

		repository.NewProblemRepository,
		service.NewProblemService,

		web.NewProblemHandler,

		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
