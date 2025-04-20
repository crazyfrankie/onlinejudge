//go:build wireinject

package user

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/crazyfrankie/onlinejudge/internal/auth"
	"github.com/crazyfrankie/onlinejudge/internal/sm"
	"github.com/crazyfrankie/onlinejudge/internal/user/repository"
	"github.com/crazyfrankie/onlinejudge/internal/user/repository/cache"
	"github.com/crazyfrankie/onlinejudge/internal/user/repository/dao"
	"github.com/crazyfrankie/onlinejudge/internal/user/service"
	"github.com/crazyfrankie/onlinejudge/internal/user/service/oauth/github"
	"github.com/crazyfrankie/onlinejudge/internal/user/service/oauth/wechat"
	"github.com/crazyfrankie/onlinejudge/internal/user/web"
	"github.com/crazyfrankie/onlinejudge/internal/user/web/third"
	ratelimit2 "github.com/crazyfrankie/onlinejudge/pkg/ratelimit"
	"github.com/crazyfrankie/onlinejudge/pkg/zapx"
)

func InitModule(l *zapx.Logger, cmd redis.Cmdable, db *gorm.DB, limiter ratelimit2.Limiter, mdlModule *auth.Module, smsModule *sm.Module) *Module {
	wire.Build(
		dao.NewUserDao,
		cache.NewUserCache,

		repository.NewUserRepository,

		service.NewUserService,
		github.NewService,
		wechat.NewService,

		web.NewUserHandler,
		third.NewOAuthGithubHandler,
		third.NewOAuthWeChatHandler,

		wire.FieldsOf(new(*sm.Module), "Sm"),
		wire.FieldsOf(new(*auth.Module), "Hdl"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
