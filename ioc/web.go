package ioc

import (
	"github.com/crazyfrankie/onlinejudge/internal/article"
	"github.com/crazyfrankie/onlinejudge/internal/auth"
	"github.com/crazyfrankie/onlinejudge/internal/judgement"
	"github.com/crazyfrankie/onlinejudge/internal/problem"
	"github.com/crazyfrankie/onlinejudge/internal/user"
	ijwt "github.com/crazyfrankie/onlinejudge/internal/user/middleware/jwt"
	"github.com/crazyfrankie/onlinejudge/internal/user/middleware/ratelimit"
	"github.com/crazyfrankie/onlinejudge/internal/user/web/third"
	rate "github.com/crazyfrankie/onlinejudge/pkg/ratelimit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func InitWebServer(mdl []gin.HandlerFunc, userHdl *user.Handler, proHdl *problem.Handler, oauthHdl *third.OAuthWeChatHandler, localHdl *judgement.LocHandler, remoteHdl *judgement.RemHandler, gitHdl *third.OAuthGithubHandler, artHdl *article.Handler, adminHdl *article.AdminHandler) *gin.Engine {
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
	adminHdl.RegisterRoute(server)
	return server
}

func GinMiddlewares(limiter rate.Limiter, jwtHdl ijwt.Handler) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		//func(c *gin.Context) {
		//	c.Set("claims", ijwt.Claims{
		//		Id: 1,
		//	})
		//},
		cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:8081"}, // 允许的前端域名
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "x-jwt-token", "x-refresh-token"},
			ExposeHeaders:    []string{"Content-Length", "x-jwt-token", "x-refresh-token"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),

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
			//IgnorePaths("/remote/run").
			//IgnorePaths("/articles/edit").
			//IgnorePaths("/remote/submit").
			//IgnorePaths("/local/run").
			AdminPaths("/admin/problem").
			AdminPaths("/admin/problem/update").
			AdminPaths("/admin/tags/create").
			AdminPaths("/admin/tags/modify").
			AdminPaths("/admin/tags").
			CheckLogin(),
	}
}
