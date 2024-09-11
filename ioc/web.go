package ioc

import (
	"github.com/gin-gonic/gin"

	"oj/internal/middleware"
	pwb "oj/internal/problem/web"
	uwb "oj/internal/user/web"
	"oj/internal/user/web/pkg/middlewares/ratelimit"
	rate "oj/pkg/ratelimit"
)

func InitWebServer(mdl []gin.HandlerFunc, userHdl *uwb.UserHandler, proHdl *pwb.ProblemHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdl...)
	// 注册路由
	userHdl.RegisterRoute(server)
	proHdl.RegisterRoute(server)
	return server
}

func GinMiddlewares(limiter rate.Limiter) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.CORS(),

		ratelimit.NewBuilder(limiter).Build(),

		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePaths("/user/signup").
			IgnorePaths("/user/signup/send-code").
			IgnorePaths("/user/signup/verify-code").
			IgnorePaths("/user/login").
			IgnorePaths("/user/login/send-code").
			IgnorePaths("/user/login-sms").
			CheckLogin(),

		middleware.NewProblemJWTMiddlewareBuilder().
			SecretPaths("/admin/problem/create").
			SecretPaths("/admin/problem").
			SecretPaths("/admin/problem/update").
			SecretPaths("/admin/tags/create").
			SecretPaths("/admin/tags/modify").
			SecretPaths("/admin/tags").
			CheckLogin(),
	}
}
