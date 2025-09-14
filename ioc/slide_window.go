package ioc

import (
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/onlinejudge/infra/contract/ratelimit"
	ratelimitimpl "github.com/crazyfrankie/onlinejudge/infra/impl/ratelimit"
)

func InitSlideWindow(cmd redis.Cmdable) ratelimit.Limiter {
	return ratelimitimpl.NewRedisSlideWindowLimiter(cmd, time.Second, 3000)
}
