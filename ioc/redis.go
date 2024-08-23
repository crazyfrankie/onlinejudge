package ioc

import (
	"github.com/redis/go-redis/v9"

	"oj/config"
)

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.Addr,
		Password: "",
	})
}
