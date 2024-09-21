package ioc

import (
	"github.com/google/wire"
	"oj/internal/middleware"
	"oj/internal/user/repository"
	"oj/internal/user/repository/cache"
	"oj/internal/user/repository/dao"
	"oj/internal/user/service"
	"oj/internal/user/web"
)

var UserSet = wire.NewSet(
	dao.NewUserDao,
	cache.NewUserCache,
	cache.NewRedisCodeCache,

	repository.NewUserRepository,
	repository.NewCodeRepository,

	service.NewUserService,
	service.NewCodeService,
	InitWechatService,
	InitSMSService,

	middleware.NewJWTService,

	web.NewUserHandler,
	web.NewOAuthHandler,
)
