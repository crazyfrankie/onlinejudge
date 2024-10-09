package ioc

import (
	"github.com/google/wire"
	"oj/internal/judgement/repository"
	"oj/internal/judgement/repository/cache"
	"oj/internal/judgement/service/local"
	"oj/internal/judgement/web"
)

var JudgeSet = wire.NewSet(
	cache.NewSubmitCache,
	cache.NewLocalSubmitCache,

	repository.NewSubmitRepository,
	repository.NewLocalSubmitRepo,

	InitJudgeService,
	local.NewLocSubmitService,

	web.NewSubmissionHandler,
	web.NewLocalSubmitHandler,
)
