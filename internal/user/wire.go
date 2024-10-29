//go:build wireinject

package user

import (
	"os"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"

	"oj/internal/user/middleware/jwt"
	"oj/internal/user/repository"
	"oj/internal/user/repository/cache"
	"oj/internal/user/repository/dao"
	"oj/internal/user/service"
	"oj/internal/user/service/oauth/github"
	"oj/internal/user/service/oauth/wechat"
	"oj/internal/user/service/sms"
	"oj/internal/user/service/sms/failover"
	"oj/internal/user/service/sms/memory"
	"oj/internal/user/service/sms/ratelimit"
	"oj/internal/user/web"
	"oj/internal/user/web/third"
	ratelimit2 "oj/pkg/ratelimit"
	"oj/pkg/zapx"
)

func InitWechatService() wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("environment variable appId not found")
	}

	var appKey string
	appKey, ok = os.LookupEnv("WECHAT_APP_KEY")
	if !ok {
		panic("environment variable appKey not found")
	}
	return wechat.NewService(appId, appKey)
}

func InitSMSService(limiter ratelimit2.Limiter) sms.Service {
	memoryService := memory.NewService()

	services := []sms.Service{
		memoryService,
	}
	rateLimitService := ratelimit.NewService(memoryService, limiter)
	failOverService := failover.NewFailOver(append(services, rateLimitService))
	return failOverService
}

func InitGithubService() github.Service {
	return github.NewService()
}

func InitLogger() *zap.Logger {
	encodeConfig := zap.NewDevelopmentEncoderConfig()
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encodeConfig), zapcore.AddSync(os.Stdout), zapcore.DebugLevel)

	customCore := zapx.NewCustomCore(core)
	logger := zap.New(customCore)

	return logger
}

func InitSlideWindow(cmd redis.Cmdable) ratelimit2.Limiter {
	return ratelimit2.NewRedisSlideWindowLimiter(cmd, time.Second, 3000)
}

func InitUserService(cmd redis.Cmdable, db *gorm.DB) service.UserService {
	wire.Build(
		dao.NewUserDao,
		cache.NewUserCache,

		repository.NewUserRepository,

		service.NewUserService,
	)
	return new(service.UserSvc)
}

func InitUserHandler(cmd redis.Cmdable, db *gorm.DB) *web.UserHandler {
	wire.Build(
		cache.NewRedisCodeCache,

		repository.NewCodeRepository,

		InitSlideWindow,

		InitUserService,
		service.NewCodeService,
		InitSMSService,

		jwt.NewRedisJWTHandler,

		web.NewUserHandler,

		InitLogger,
	)
	return new(web.UserHandler)
}

func InitOAuthGithubHandler(cmd redis.Cmdable, db *gorm.DB) *third.OAuthGithubHandler {
	wire.Build(
		InitUserService,
		InitGithubService,
		jwt.NewRedisJWTHandler,
		third.NewOAuthGithubHandler,
	)
	return new(third.OAuthGithubHandler)
}

func InitOAuthWeChatHandler(cmd redis.Cmdable, db *gorm.DB) *third.OAuthWeChatHandler {
	wire.Build(
		InitUserService,
		jwt.NewRedisJWTHandler,
		InitWechatService,
		third.NewOAuthHandler,
	)
	return new(third.OAuthWeChatHandler)
}
