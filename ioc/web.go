package ioc

import (
	"github.com/gin-gonic/gin"

	"oj/internal/article"
	"oj/internal/auth"
	"oj/internal/judgement"
	"oj/internal/problem"
	"oj/internal/user"
	ijwt "oj/internal/user/middleware/jwt"
	"oj/internal/user/middleware/ratelimit"
	"oj/internal/user/web/third"
	rate "oj/pkg/ratelimit"
)

func InitWebServer(mdl []gin.HandlerFunc, userHdl *user.Handler, proHdl *problem.Handler, oauthHdl *third.OAuthWeChatHandler, localHdl *judgement.LocHandler, remoteHdl *judgement.RemHandler, gitHdl *third.OAuthGithubHandler, artHdl *article.Handler) *gin.Engine {
	server := gin.Default()
	server.Use(mdl...)
	// 注册路由
	userHdl.RegisterRoute(server)
	proHdl.RegisterRoute(server)
	oauthHdl.RegisterRoute(server)
	localHdl.RegisterRoute(server)
	remoteHdl.RegisterRoute(server)
	gitHdl.RegisterRoute(server)
	artHdl.RegisterRoute(server)
	return server
}

func GinMiddlewares(limiter rate.Limiter, jwtHdl ijwt.Handler) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		auth.CORS(),

		ratelimit.NewBuilder(limiter).Build(),

		auth.NewLoginJWTMiddlewareBuilder(jwtHdl).
			IgnorePaths("/user/signup").
			IgnorePaths("/user/login").
			IgnorePaths("/user/send-code").
			IgnorePaths("/user/verify-code").
			IgnorePaths("/oauth/wechat/url").
			IgnorePaths("/oauth/github/url").
			IgnorePaths("/oauth/github/callback").
			IgnorePaths("/user/refresh-token").
			IgnorePaths("/remote/run").
			//IgnorePaths("/remote/submit").
			IgnorePaths("/local/run").
			AdminPaths("/admin/problem").
			AdminPaths("/admin/problem/update").
			AdminPaths("/admin/tags/create").
			AdminPaths("/admin/tags/modify").
			AdminPaths("/admin/tags").
			CheckLogin(),
	}
}
