package web

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"oj/config"
	"oj/user/repository"
	"oj/user/repository/cache"
	"oj/user/repository/dao"
	"oj/user/service"
	"oj/user/service/sms/memory"
	"oj/user/web/middleware"
	"oj/user/web/pkg/middlewares/ratelimit"
)

// InitUserHandler Handler 对象的创建
func InitUserHandler(redisClient redis.Cmdable) *UserHandler {
	db := dao.InitDB()
	db.AutoMigrate(&dao.User{})

	codeCache := cache.NewCodeCache(redisClient)
	codeRepo := repository.NewCodeRepository(codeCache)
	smsSvc := memory.NewService()
	codeSvc := service.NewCodeService(codeRepo, smsSvc)

	ud := dao.NewUserDao(db)
	ce := cache.NewUserCache(redisClient, time.Minute*15)
	repo := repository.NewUserRepository(ud, ce)
	svc := service.NewUserService(repo)
	u := NewUserHandler(svc, codeSvc)

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

	// 限流
	router.Use(ratelimit.NewBuilder(rateLimitCmd, time.Second, 100).Build())

	// 登录校验
	router.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/user/signup").
		IgnorePaths("/user/login").
		IgnorePaths("/user/login_sms/code/send").
		IgnorePaths("/user/sms_login").
		CheckLogin())

	// 路由注册
	u := InitUserHandler(userCacheCmd)
	u.RegisterRoute(router)

	return router
}
