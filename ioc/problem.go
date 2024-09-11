package ioc

import (
	"github.com/google/wire"
	"oj/internal/problem/repository"
	"oj/internal/problem/repository/cache"
	"oj/internal/problem/repository/dao"
	"oj/internal/problem/service"
	"oj/internal/problem/web"
)

var ProblemSet = wire.NewSet(

	dao.NewProblemDao,
	cache.NewProblemCache,

	repository.NewProblemRepository,

	service.NewProblemService,

	web.NewProblemHandler,
)
