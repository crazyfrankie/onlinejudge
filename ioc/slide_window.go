package ioc

import (
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/onlinejudge/pkg/ratelimit"
)

func InitSlideWindow(cmd redis.Cmdable) ratelimit.Limiter {
	return ratelimit.NewRedisSlideWindowLimiter(cmd, time.Second, 3000)
}
