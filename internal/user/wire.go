//go:build wireinject

package user

import (
	"github.com/crazyfrankie/onlinejudge/internal/auth"
	"github.com/crazyfrankie/onlinejudge/internal/user/service/sms/failover"
	"github.com/crazyfrankie/onlinejudge/internal/user/service/sms/memory"
	"github.com/crazyfrankie/onlinejudge/internal/user/service/sms/ratelimit"
	"github.com/crazyfrankie/onlinejudge/internal/user/web"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/crazyfrankie/onlinejudge/internal/user/repository"
	"github.com/crazyfrankie/onlinejudge/internal/user/repository/cache"
	"github.com/crazyfrankie/onlinejudge/internal/user/repository/dao"
	"github.com/crazyfrankie/onlinejudge/internal/user/service"
	"github.com/crazyfrankie/onlinejudge/internal/user/service/oauth/github"
	"github.com/crazyfrankie/onlinejudge/internal/user/service/oauth/wechat"
	"github.com/crazyfrankie/onlinejudge/internal/user/service/sms"
	"github.com/crazyfrankie/onlinejudge/internal/user/web/third"
	ratelimit2 "github.com/crazyfrankie/onlinejudge/pkg/ratelimit"
)

//func InitLogger() *zap.Logger {
//	encodeConfig := zap.NewDevelopmentEncoderConfig()
//	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encodeConfig), zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
//
//	customCore := zapx.NewCustomCore(core)
//	logger := zap.New(customCore)
//
//	return logger
//}

func InitSMSService(limiter ratelimit2.Limiter) sms.Service {
	memoryService := memory.NewService()

	services := []sms.Service{
		memoryService,
	}
	rateLimitService := ratelimit.NewService(memoryService, limiter)
	failOverService := failover.NewFailOver(append(services, rateLimitService))
	return failOverService
}

func InitModule(cmd redis.Cmdable, db *gorm.DB, limiter ratelimit2.Limiter, module *auth.Module) *Module {
	wire.Build(
		dao.NewUserDao,
		cache.NewUserCache,
		cache.NewRedisCodeCache,

		repository.NewCodeRepository,
		repository.NewUserRepository,

		service.NewUserService,
		service.NewCodeService,
		github.NewService,
		wechat.NewService,
		InitSMSService,

		web.NewUserHandler,
		third.NewOAuthGithubHandler,
		third.NewOAuthWeChatHandler,

		wire.FieldsOf(new(*auth.Module), "Hdl"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
