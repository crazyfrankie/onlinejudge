//go:build wireinject

package ioc

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"oj/internal/article"
	"oj/internal/judgement"
	"oj/internal/problem"
	"oj/internal/user"
	"oj/internal/user/middleware/jwt"
)

var BaseSet = wire.NewSet(InitDB, InitRedis)

func InitGin() *gin.Engine {
	wire.Build(
		BaseSet,

		user.InitUserHandler,
		user.InitOAuthGithubHandler,
		user.InitOAuthWeChatHandler,

		problem.InitProblemHandler,

		judgement.InitLocalJudgement,
		judgement.InitRemoteJudgement,

		article.InitArticleHandler,

		jwt.NewRedisJWTHandler,

		InitSlideWindow,
		// gin 的中间件
		GinMiddlewares,

		// web 服务器
		InitWebServer,
	)
	return new(gin.Engine)
}
