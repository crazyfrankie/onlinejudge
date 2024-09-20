package ioc

import (
	"github.com/redis/go-redis/v9"
	"oj/pkg/ratelimit"
	"time"
)

func NewLimiter(cmd redis.Cmdable) ratelimit.Limiter {
	return ratelimit.NewRedisSlideWindowLimiter(cmd, time.Second, 3000)
}
