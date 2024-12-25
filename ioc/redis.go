package ioc

import (
	"time"

	"github.com/crazyfrankie/onlinejudge/config"
	"github.com/redis/go-redis/v9"
)

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr:         config.GetConf().Redis.Address,
		Password:     "",
		MinIdleConns: config.GetConf().Redis.MinIdleConns,
		PoolSize:     config.GetConf().Redis.PoolSize,
		DialTimeout:  time.Minute * 5,
	})
}
