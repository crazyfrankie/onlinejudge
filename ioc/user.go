package ioc

import (
	"github.com/google/wire"

	"oj/internal/user/middleware/jwt"
	"oj/internal/user/repository"
	"oj/internal/user/repository/cache"
	"oj/internal/user/repository/dao"
	"oj/internal/user/service"
	"oj/internal/user/web"
	"oj/internal/user/web/third"
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
	InitGithubService,

	jwt.NewRedisJWTHandler,

	web.NewUserHandler,
	third.NewOAuthHandler,
	third.NewOAuthGithubHandler,

	InitLogger,
)
