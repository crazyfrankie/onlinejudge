package ioc

import (
	"github.com/gin-gonic/gin"

	jwb "oj/internal/judgement/web"
	pwb "oj/internal/problem/web"
	uwb "oj/internal/user/web"
	ijwt "oj/internal/user/web/jwt"
	"oj/internal/user/web/middlewares"
	"oj/internal/user/web/middlewares/ratelimit"
	rate "oj/pkg/ratelimit"
)

func InitWebServer(mdl []gin.HandlerFunc, userHdl *uwb.UserHandler, proHdl *pwb.ProblemHandler, oauthHdl *uwb.OAuthWeChatHandler, judgeHdl *jwb.SubmissionHandler, localHdl *jwb.LocalSubmitHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdl...)
	// 注册路由
	userHdl.RegisterRoute(server)
	proHdl.RegisterRoute(server)
	oauthHdl.RegisterRoute(server)
	judgeHdl.RegisterRoute(server)
	localHdl.RegisterRoute(server)
	return server
}

func GinMiddlewares(limiter rate.Limiter, jwtHdl ijwt.Handler) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middlewares.CORS(),

		ratelimit.NewBuilder(limiter).Build(),

		middlewares.NewLoginJWTMiddlewareBuilder(jwtHdl).
			IgnorePaths("/user/signup").
			IgnorePaths("/user/signup/send-code").
			IgnorePaths("/user/signup/verify-code").
			IgnorePaths("/user/login").
			IgnorePaths("/user/login/send-code").
			IgnorePaths("/user/login-sms").
			IgnorePaths("/oauth/wechat/authurl").
			IgnorePaths("/user/refresh-token").
			IgnorePaths("/remote/run").
			//IgnorePaths("/remote/submit").
			IgnorePaths("/local/run").
			CheckLogin(),

		middlewares.NewProblemJWTMiddlewareBuilder(jwtHdl).
			//SecretPaths("/admin/problem/create").
			SecretPaths("/admin/problem").
			SecretPaths("/admin/problem/update").
			SecretPaths("/admin/tags/create").
			SecretPaths("/admin/tags/modify").
			SecretPaths("/admin/tags").
			CheckLogin(),
	}
}
