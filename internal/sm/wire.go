//go:build wireinject

package sm

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/onlinejudge/internal/sm/repository"
	"github.com/crazyfrankie/onlinejudge/internal/sm/repository/cache"
	svc "github.com/crazyfrankie/onlinejudge/internal/sm/service"
	"github.com/crazyfrankie/onlinejudge/internal/sm/service/sms/service"
	"github.com/crazyfrankie/onlinejudge/internal/sm/service/sms/service/failover"
	"github.com/crazyfrankie/onlinejudge/internal/sm/service/sms/service/memory"
	"github.com/crazyfrankie/onlinejudge/internal/sm/service/sms/service/metric"
	"github.com/crazyfrankie/onlinejudge/internal/sm/service/sms/service/ratelimit"
	ratelimit2 "github.com/crazyfrankie/onlinejudge/pkg/ratelimit"
)

func InitSMS(limiter ratelimit2.Limiter) service.Service {
	memoryService := memory.NewService()
	services := []service.Service{
		memoryService,
	}
	rateLimitService := ratelimit.NewService(memoryService, limiter)
	failOverService := failover.NewFailOver(append(services, rateLimitService))
	// 接入监控
	metricService := metric.NewProemtheusDecortor(failOverService)

	return metricService
}

func InitModule(cmd redis.Cmdable, limiter ratelimit2.Limiter) *Module {
	wire.Build(
		cache.NewRedisCodeCache,
		repository.NewCodeRepository,
		InitSMS,
		svc.NewCodeService,

		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
