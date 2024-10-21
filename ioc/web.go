package ioc

import (
	"github.com/gin-gonic/gin"

	jwb "oj/internal/judgement/web"
	pwb "oj/internal/problem/web"
	ijwt "oj/internal/user/middleware/jwt"
	"oj/internal/user/middleware/ratelimit"
	uwb "oj/internal/user/web"
	"oj/internal/user/web/auth"
	"oj/internal/user/web/third"
	rate "oj/pkg/ratelimit"
)

func InitWebServer(mdl []gin.HandlerFunc, userHdl *uwb.UserHandler, proHdl *pwb.ProblemHandler, oauthHdl *third.OAuthWeChatHandler, judgeHdl *jwb.SubmissionHandler, localHdl *jwb.LocalSubmitHandler, gitHdl *third.OAuthGithubHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdl...)
	// 注册路由
	userHdl.RegisterRoute(server)
	proHdl.RegisterRoute(server)
	oauthHdl.RegisterRoute(server)
	judgeHdl.RegisterRoute(server)
	localHdl.RegisterRoute(server)
	gitHdl.RegisterRoute(server)
	return server
}

func GinMiddlewares(limiter rate.Limiter, jwtHdl ijwt.Handler) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		auth.CORS(),

		ratelimit.NewBuilder(limiter).Build(),

		auth.NewLoginJWTMiddlewareBuilder(jwtHdl).
			IgnorePaths("/user/signup").
			IgnorePaths("/user/send-code").
			IgnorePaths("/user/signup/verify-code").
			IgnorePaths("/user/login/verify-code").
			IgnorePaths("/user/login").
			IgnorePaths("/oauth/wechat/url").
			IgnorePaths("/oauth/github/url").
			IgnorePaths("/oauth/github/callback").
			IgnorePaths("/user/refresh-token").
			IgnorePaths("/remote/run").
			//IgnorePaths("/remote/submit").
			IgnorePaths("/local/run").
			CheckLogin(),

		auth.NewProblemJWTMiddlewareBuilder(jwtHdl).
			//SecretPaths("/admin/problem/create").
			SecretPaths("/admin/problem").
			SecretPaths("/admin/problem/update").
			SecretPaths("/admin/tags/create").
			SecretPaths("/admin/tags/modify").
			SecretPaths("/admin/tags").
			CheckLogin(),
	}
}
