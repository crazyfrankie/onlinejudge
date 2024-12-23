//go:build wireinject

package ioc

import (
	"github.com/google/wire"
	"oj/internal/article"
	"oj/internal/judgement"
	"oj/internal/problem"
	"oj/internal/user"
	"oj/internal/user/middleware/jwt"
)

var BaseSet = wire.NewSet(InitDB, InitRedis, InitKafka, InitLog)

func InitApp() *App {
	wire.Build(
		BaseSet,

		user.InitUserHandler,
		user.InitOAuthGithubHandler,
		user.InitOAuthWeChatHandler,

		problem.InitProblemHandler,

		judgement.InitLocalJudgement,
		judgement.InitRemoteJudgement,

		article.InitArticleHandler,
		article.InitConsumer,

		jwt.NewRedisJWTHandler,

		InitSlideWindow,
		// gin 的中间件
		GinMiddlewares,

		// web 服务器
		InitWebServer,


		NewConsumers,

		InitOJ,
	)
	
	return new(App)
}
