//go:build wireinject

package ioc

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"oj/user/repository"
	"oj/user/repository/cache"
	"oj/user/repository/dao"
	"oj/user/service"
	"oj/user/web"
)

func InitGinWithRedis() *gin.Engine {
	wire.Build(
		// 最底层的第三方依赖
		InitDB, InitRedis,

		dao.NewUserDao,

		cache.NewUserCache,
		cache.NewRedisCodeCache,

		repository.NewCodeRepository,
		repository.NewUserRepository,

		service.NewUserService,
		service.NewCodeService,
		InitSMSService,

		web.NewUserHandler,

		// gin 的中间件
		GinMiddlewares,

		// web 服务器
		InitWebServer,
	)
	return new(gin.Engine)
}

func InitGinWithMem() *gin.Engine {
	wire.Build(
		// 最底层的第三方依赖
		InitDB, InitRedis, InitGoMem,

		dao.NewUserDao,

		cache.NewUserCache,
		cache.NewMemCodeCache,

		repository.NewCodeRepository,
		repository.NewUserRepository,

		service.NewUserService,
		service.NewCodeService,
		InitSMSService,

		web.NewUserHandler,

		// gin 的中间件
		GinMiddlewares,

		// web 服务器
		InitWebServer,
	)
	return new(gin.Engine)
}
