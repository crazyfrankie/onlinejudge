// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

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

// Injectors from wire.go:

func InitModule(cmd redis.Cmdable, db *gorm.DB, module *problem.Module) *Module {
	localSubmitCache := cache.NewLocalSubmitCache(cmd)
	submitDao := dao.NewSubmitDao(db)
	localSubmitRepo := repository.NewLocalSubmitRepo(localSubmitCache, submitDao)
	problemRepository := module.Repo
	locSubmitService := local.NewLocSubmitService(localSubmitRepo, problemRepository)
	localSubmitHandler := web.NewLocalSubmitHandler(locSubmitService)
	submitCache := cache.NewSubmitCache(cmd)
	submitRepository := repository.NewSubmitRepository(submitCache)
	string2 := InitJudgeKey(cmd, db)
	submitService := remote.NewSubmitService(submitRepository, problemRepository, string2)
	submissionHandler := web.NewSubmissionHandler(submitService)
	judgementModule := &Module{
		LocHdl: localSubmitHandler,
		RemHdl: submissionHandler,
	}
	return judgementModule
}

// wire.go:

var LocalSet = wire.NewSet(dao.NewSubmitDao, cache.NewLocalSubmitCache, repository.NewLocalSubmitRepo, local.NewLocSubmitService, web.NewLocalSubmitHandler)

var RemoteSet = wire.NewSet(cache.NewSubmitCache, repository.NewSubmitRepository, remote.NewSubmitService, web.NewSubmissionHandler)

func InitJudgeKey(cmd redis.Cmdable, db *gorm.DB) string {
	key, ok := os.LookupEnv("RAPIDAPI_KEY")
	if !ok {
		panic("environment variable rapidapiKey not found")
	}

	return key
}
