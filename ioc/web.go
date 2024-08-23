package ioc

import (
	"strings"
	"time"

	"github.com/gin-gonic/contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"oj/user/web"
	"oj/user/web/middleware"
	"oj/user/web/pkg/middlewares/ratelimit"
)

func InitWebServer(mdl []gin.HandlerFunc, userHdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdl...)
	// 注册路由
	userHdl.RegisterRoute(server)
	return server
}

func GinMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowedHeaders: []string{"Content-Type", "Authorization"},
			// 是否允许带 cookie 之类的
			AllowCredentials: true,
			AllowOriginFunc: func(origin string) bool {
				if strings.HasPrefix(origin, "localhost") {
					// 开发环境
					return true
				}
				return strings.Contains(origin, "yourcompany.com")
			},
			// 不加这一行 前端拿不到 token
			ExposedHeaders: []string{"x-jwt-token"},
			MaxAge:         12 * time.Hour,
		}),

		ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),

		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePaths("/user/signup").
			IgnorePaths("/user/login").
			IgnorePaths("/user/login_sms/code/send").
			IgnorePaths("/user/sms_login").
			CheckLogin(),
	}
}
