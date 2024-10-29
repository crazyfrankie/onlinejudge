//go:build wireinject

package judgement

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"oj/internal/judgement/repository"
	"oj/internal/judgement/repository/cache"
	"oj/internal/judgement/service/local"
	"oj/internal/judgement/service/remote"
	"oj/internal/judgement/web"
	"oj/internal/problem"
	"os"
)

var LocalSet = wire.NewSet(
	cache.NewLocalSubmitCache,

	repository.NewLocalSubmitRepo,

	problem.InitProblemRepo,

	local.NewLocSubmitService,

	web.NewLocalSubmitHandler,
)

var RemoteSet = wire.NewSet(
	cache.NewSubmitCache,

	repository.NewSubmitRepository,

	problem.InitProblemRepo,

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

func InitLocalJudgement(cmd redis.Cmdable, db *gorm.DB) *web.LocalSubmitHandler {
	wire.Build(
		LocalSet,
	)
	return new(web.LocalSubmitHandler)
}

func InitRemoteJudgement(cmd redis.Cmdable, db *gorm.DB) *web.SubmissionHandler {
	wire.Build(
		InitJudgeKey,
		RemoteSet,
	)
	return new(web.SubmissionHandler)
}
