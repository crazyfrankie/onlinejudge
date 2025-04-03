//go:build wireinject

package judgement

import (
	"github.com/crazyfrankie/onlinejudge/internal/judgement/repository"
	"github.com/crazyfrankie/onlinejudge/internal/judgement/repository/cache"
	"github.com/crazyfrankie/onlinejudge/internal/judgement/repository/dao"
	"github.com/crazyfrankie/onlinejudge/internal/judgement/service/local"
	"github.com/crazyfrankie/onlinejudge/internal/judgement/service/remote"
	"github.com/crazyfrankie/onlinejudge/internal/judgement/web"
	"github.com/crazyfrankie/onlinejudge/internal/problem"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"os"
)

var LocalSet = wire.NewSet(
	dao.NewSubmitDao,
	cache.NewLocalSubmitCache,
	repository.NewLocalSubmitRepo,

	local.NewLocSubmitService,

	web.NewLocalSubmitHandler,
)

var RemoteSet = wire.NewSet(
	cache.NewSubmitCache,

	repository.NewSubmitRepository,

	remote.NewSubmitService,

	web.NewSubmissionHandler,
)

func InitJudgeKey(cmd redis.Cmdable, db *gorm.DB) string {
	key, ok := os.LookupEnv("RAPIDAPI_KEY")
	if !ok {
		panic("environment variable rapidapiKey not found")
	}

	return key
}

func InitModule(cmd redis.Cmdable, db *gorm.DB, module *problem.Module) *Module {
	wire.Build(
		LocalSet,
		RemoteSet,
		InitJudgeKey,

		wire.FieldsOf(new(*problem.Module), "Repo"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
