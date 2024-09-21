package ioc

import (
	"github.com/google/wire"
	"oj/internal/judgement/repository"
	"oj/internal/judgement/repository/cache"
	"oj/internal/judgement/web"
)

var JudgeSet = wire.NewSet(
	cache.NewSubmitCache,

	repository.NewSubmitRepository,

	InitJudgeService,

	web.NewSubmissionHandler,
)
