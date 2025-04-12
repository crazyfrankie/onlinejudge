package ioc

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/crazyfrankie/onlinejudge/common/response"
	"github.com/crazyfrankie/onlinejudge/internal/article"
	"github.com/crazyfrankie/onlinejudge/internal/auth"
	"github.com/crazyfrankie/onlinejudge/internal/judgement"
	"github.com/crazyfrankie/onlinejudge/internal/mws"
	"github.com/crazyfrankie/onlinejudge/internal/problem"
	"github.com/crazyfrankie/onlinejudge/internal/user"
	"github.com/crazyfrankie/onlinejudge/internal/user/web/third"
	rate "github.com/crazyfrankie/onlinejudge/pkg/ratelimit"
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

func GinMiddlewares(cmd redis.Cmdable, limiter rate.Limiter, jwtHdl auth.JWTHandler, authz mws.Authorizer) []gin.HandlerFunc {
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

		(&mws.MetricsBuilder{
			Namespace: "cfc_studio_frank",
			Subsystem: "onlinejudge",
			Name:      "gin_http",
			Help:      "统计 Gin 的 HTTP 接口",
		}).Builder(),

		otelgin.Middleware("onlinejudge"),

		mws.NewBuilder(limiter).Build(),

		mws.NewAuthnHandler(cmd, jwtHdl).
			IgnorePaths("/api/user/login").
			IgnorePaths("/api/user/send-code").
			IgnorePaths("/api/user/verify-code").
			IgnorePaths("/api/user/refresh-token").
			IgnorePaths("/api/oauth/wechat/url").
			IgnorePaths("/api/oauth/github/url").
			IgnorePaths("/api/oauth/github/callback").
			IgnorePaths("/api/oauth/wechat/callback").
			IgnorePaths("/api/user/test").
			Authn(),

		mws.NewAuthzHandler(authz).Authz(),
	}
}
