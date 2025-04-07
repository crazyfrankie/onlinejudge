package ioc

import (
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/onlinejudge/config"
)

func InitRedis() redis.Cmdable {
	client := redis.NewClient(&redis.Options{
		Addr:         config.GetConf().Redis.Address,
		Password:     "",
		MinIdleConns: config.GetConf().Redis.MinIdleConns,
		PoolSize:     config.GetConf().Redis.PoolSize,
		DialTimeout:  time.Minute * 5,
	})

	// tracing instrumentation
	if err := redisotel.InstrumentTracing(client); err != nil {
		panic(err)
	}

	// metrics instrumentation.
	if err := redisotel.InstrumentMetrics(client); err != nil {
		panic(err)
	}

	return client
}
