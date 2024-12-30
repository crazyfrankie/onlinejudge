package ioc

import (
	"github.com/crazyfrankie/onlinejudge/common/response"
	"github.com/crazyfrankie/onlinejudge/internal/article"
	"github.com/crazyfrankie/onlinejudge/internal/judgement"
	"github.com/crazyfrankie/onlinejudge/internal/middleware/auth"
	ijwt "github.com/crazyfrankie/onlinejudge/internal/middleware/jwt"
	"github.com/crazyfrankie/onlinejudge/internal/middleware/metrics"
	"github.com/crazyfrankie/onlinejudge/internal/middleware/ratelimit"
	"github.com/crazyfrankie/onlinejudge/internal/problem"
	"github.com/crazyfrankie/onlinejudge/internal/user"
	"github.com/crazyfrankie/onlinejudge/internal/user/web"
	"github.com/crazyfrankie/onlinejudge/internal/user/web/third"
	rate "github.com/crazyfrankie/onlinejudge/pkg/ratelimit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
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
	(&web.MetricHandler{}).RegisterRoute(server)
	return server
}

func GinMiddlewares(limiter rate.Limiter, jwtHdl ijwt.Handler) []gin.HandlerFunc {
	response.InitCouter(prometheus.CounterOpts{
		Namespace: "cfc_studio_frank",
		Subsystem: "onlinejudge",
		Name:      "http_biz_code",
		Help:      "HTTP 的业务错误码",
	})
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:8081"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length", "x-jwt-token", "x-refresh-token"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),

		(&metrics.MetricsBuilder{
			Namespace: "cfc_studio_frank",
			Subsystem: "onlinejudge",
			Name:      "gin_http",
			Help:      "统计 Gin 的 HTTP 接口",
		}).Builder(),

		ratelimit.NewBuilder(limiter).Build(),

		auth.NewLoginJWTMiddlewareBuilder(jwtHdl).
			IgnorePaths("/api/user/login").
			IgnorePaths("/api/user/send-code").
			IgnorePaths("/api/user/verify-code").
			IgnorePaths("/api/user/refresh-token").
			IgnorePaths("/api/oauth/wechat/url").
			IgnorePaths("/api/oauth/github/url").
			IgnorePaths("/api/oauth/github/callback").
			IgnorePaths("/api/oauth/wechat/callback").
			IgnorePaths("/api/test/metric").
			AdminPaths("/api/admin/problem").
			AdminPaths("/api/admin/problem/update").
			AdminPaths("/api/admin/tags/create").
			AdminPaths("/api/admin/tags/modify").
			AdminPaths("/api/admin/tags").
			CheckLogin(),
	}
}
