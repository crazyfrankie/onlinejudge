//go:build wireinject

package user

import (
	"github.com/crazyfrankie/onlinejudge/internal/middleware"
	"github.com/crazyfrankie/onlinejudge/internal/sms"
	"github.com/crazyfrankie/onlinejudge/internal/user/web"
	"go.uber.org/zap"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/crazyfrankie/onlinejudge/internal/user/repository"
	"github.com/crazyfrankie/onlinejudge/internal/user/repository/cache"
	"github.com/crazyfrankie/onlinejudge/internal/user/repository/dao"
	"github.com/crazyfrankie/onlinejudge/internal/user/service"
	"github.com/crazyfrankie/onlinejudge/internal/user/service/oauth/github"
	"github.com/crazyfrankie/onlinejudge/internal/user/service/oauth/wechat"
	"github.com/crazyfrankie/onlinejudge/internal/user/web/third"
	ratelimit2 "github.com/crazyfrankie/onlinejudge/pkg/ratelimit"
)

func InitModule(l *zap.Logger, cmd redis.Cmdable, db *gorm.DB, limiter ratelimit2.Limiter, mdlModule *middleware.Module, smsModule *sms.Module) *Module {
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

		web.NewUserHandler,
		third.NewOAuthGithubHandler,
		third.NewOAuthWeChatHandler,

		wire.FieldsOf(new(*sms.Module), "SmsSvc"),
		wire.FieldsOf(new(*middleware.Module), "Hdl"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
