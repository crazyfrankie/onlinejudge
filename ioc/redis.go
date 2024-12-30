package ioc

import (
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/onlinejudge/config"
	"github.com/crazyfrankie/onlinejudge/pkg/redisx"
	"github.com/prometheus/client_golang/prometheus"
)

func InitRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         config.GetConf().Redis.Address,
		Password:     "",
		MinIdleConns: config.GetConf().Redis.MinIdleConns,
		PoolSize:     config.GetConf().Redis.PoolSize,
		DialTimeout:  time.Minute * 5,
	})

	client.AddHook(redisx.NewPrometheusHook(prometheus.SummaryOpts{
		Namespace: "cfcstudio_frank",
		Subsystem: "onlinejudge",
		Name:      "redis_resp_time",
		Help:      "统计 Redis 的执行时间",
	}))

	return client
}
