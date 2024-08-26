package ioc

import (
	"github.com/redis/go-redis/v9"
	"oj/config"
	"time"
)

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr:         config.Config.Redis.Addr,
		Password:     "",
		MinIdleConns: config.Config.Redis.MinIdleConns,
		PoolSize:     config.Config.Redis.PoolSize,
		DialTimeout:  time.Minute * 5,
	})
}
