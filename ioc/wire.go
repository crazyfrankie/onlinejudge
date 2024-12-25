//go:build wireinject

package ioc

import (
	"github.com/crazyfrankie/onlinejudge/internal/article"
	"github.com/crazyfrankie/onlinejudge/internal/judgement"
	"github.com/crazyfrankie/onlinejudge/internal/problem"
	"github.com/crazyfrankie/onlinejudge/internal/user"
	"github.com/google/wire"
)

var BaseSet = wire.NewSet(InitDB, InitRedis, InitKafka, InitLog)

func InitApp() *App {
	wire.Build(
		BaseSet,

		user.InitModule,

		problem.InitModule,

		judgement.InitModule,

		article.InitModule,

		InitSlideWindow,
		// gin 的中间件
		GinMiddlewares,

		// web 服务器
		InitWebServer,

		NewConsumers,

		wire.FieldsOf(new(*article.Module), "Consumer"),
		wire.FieldsOf(new(*user.Module), "Hdl"),
		wire.FieldsOf(new(*user.Module), "JWTHdl"),
		wire.FieldsOf(new(*user.Module), "GithubHdl"),
		wire.FieldsOf(new(*user.Module), "WeChatHdl"),
		wire.FieldsOf(new(*problem.Module), "Hdl"),
		wire.FieldsOf(new(*judgement.Module), "LocHdl"),
		wire.FieldsOf(new(*judgement.Module), "RemHdl"),
		wire.FieldsOf(new(*article.Module), "Hdl"),
		wire.FieldsOf(new(*article.Module), "AdminHdl"),
		wire.Struct(new(App), "*"),
	)

	return new(App)
}
