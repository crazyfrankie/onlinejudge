package ioc

import (
	"oj/problem/pwb"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"oj/middleware"
	"oj/user/uwb"
	"oj/user/uwb/pkg/middlewares/ratelimit"
)

func InitWebServer(mdl []gin.HandlerFunc, userHdl *uwb.UserHandler, proHdl *pwb.ProblemHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdl...)
	// 注册路由
	userHdl.RegisterRoute(server)
	proHdl.RegisterRoute(server)
	return server
}

func GinMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.CORS(),

		ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),

		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePaths("/user/signup").
			IgnorePaths("/user/signup/send-code").
			IgnorePaths("/user/signup/verify-code").
			IgnorePaths("/user/login").
			IgnorePaths("/user/login/send-code").
			IgnorePaths("/user/login-sms").
			CheckLogin(),

		middleware.NewProblemJWTMiddlewareBuilder().
			SecretPaths("/problem/create").
			SecretPaths("/problem/delete").
			SecretPaths("/problem/update").
			CheckLogin(),
	}
}
