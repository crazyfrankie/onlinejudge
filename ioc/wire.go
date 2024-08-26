//go:build wireinject

package ioc

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"oj/internal/problem/web"
)

var BaseSet = wire.NewSet(InitDB, InitRedis)

func InitGin() *gin.Engine {
	wire.Build(
		BaseSet,

		InitUserHandler,
		web.NewProblemHandler,

		// gin 的中间件
		GinMiddlewares,

		// web 服务器
		InitWebServer,
	)
	return new(gin.Engine)
}
