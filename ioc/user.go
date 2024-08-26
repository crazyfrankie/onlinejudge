package ioc

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"oj/internal/user"
	"oj/internal/user/web"
)

func InitUserHandler(db *gorm.DB, cmdable redis.Cmdable) *web.UserHandler {
	return user.InitHandler(db, cmdable)
}
