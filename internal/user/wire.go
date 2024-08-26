package user

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"oj/internal/user/repository"
	"oj/internal/user/repository/cache"
	"oj/internal/user/repository/dao"
	"oj/internal/user/service"
	"oj/internal/user/web"
)

func InitHandle(db *gorm.DB, cmd redis.Cmdable) *web.UserHandler {
	wire.Build(
		dao.NewUserDao,
		cache.NewUserCache,
		cache.NewRedisCodeCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,

		InitSMSService,
		service.NewUserService,
		service.NewCodeService,

		initHandler,
	)
	return new(web.UserHandler)
}
