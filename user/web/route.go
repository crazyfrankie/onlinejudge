package web

import (
	"oj/config"
	"oj/user/repository/cache"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"oj/user/repository"
	"oj/user/repository/dao"
	"oj/user/service"
	"oj/user/web/middleware"
	"oj/user/web/pkg/middlewares/ratelimit"
)

// InitHandler Handler 对象的创建
func InitHandler(redisClient redis.Cmdable) *UserHandler {
	db := dao.InitDB()
	db.AutoMigrate(&dao.User{})

	ud := dao.NewUserDao(db)
	ce := cache.NewUserCache(redisClient, time.Minute*15)
	repo := repository.NewUserRepository(ud, ce)
	svc := service.NewUserService(repo)
	u := NewUserHandler(svc)
	return u
}

// InitWeb gin 框架的初始化以及路由的注册
func InitWeb() *gin.Engine {
	router := gin.Default()

	// 跨域设置
	router.Use(middleware.Cors())

	// 限流 Redis 客户端
	rateLimitCmd := redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.Addr,
		Password: "",
		DB:       1,
	})

	// 用户缓存 Redis 客户端
	userCacheCmd := redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.Addr,
		Password: "",
		DB:       0,
	})

	router.Use(ratelimit.NewBuilder(rateLimitCmd, time.Second, 100).Build())

	// 登录校验
	router.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/user/signup").
		IgnorePaths("/user/login").
		CheckLogin())

	// 路由注册
	u := InitHandler(userCacheCmd)
	u.RegisterRoute(router)

	return router
}
