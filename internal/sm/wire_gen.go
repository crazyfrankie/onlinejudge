// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package sm

import (
	"github.com/crazyfrankie/onlinejudge/internal/sm/repository"
	"github.com/crazyfrankie/onlinejudge/internal/sm/repository/cache"
	"github.com/crazyfrankie/onlinejudge/internal/sm/service"
	service2 "github.com/crazyfrankie/onlinejudge/internal/sm/service/sms/service"
	"github.com/crazyfrankie/onlinejudge/internal/sm/service/sms/service/failover"
	"github.com/crazyfrankie/onlinejudge/internal/sm/service/sms/service/memory"
	"github.com/crazyfrankie/onlinejudge/internal/sm/service/sms/service/metric"
	ratelimit2 "github.com/crazyfrankie/onlinejudge/internal/sm/service/sms/service/ratelimit"
	"github.com/crazyfrankie/onlinejudge/pkg/ratelimit"
	"github.com/redis/go-redis/v9"
)

// Injectors from wire.go:

func InitModule(cmd redis.Cmdable, limiter ratelimit.Limiter) *Module {
	codeCache := cache.NewRedisCodeCache(cmd)
	codeRepository := repository.NewCodeRepository(codeCache)
	serviceService := InitSMS(limiter)
	codeService := service.NewCodeService(codeRepository, serviceService)
	module := &Module{
		Sm: codeService,
	}
	return module
}

// wire.go:

func InitSMS(limiter ratelimit.Limiter) service2.Service {
	memoryService := memory.NewService()
	services := []service2.Service{
		memoryService,
	}
	rateLimitService := ratelimit2.NewService(memoryService, limiter)
	failOverService := failover.NewFailOver(append(services, rateLimitService))

	metricService := metric.NewProemtheusDecortor(failOverService)

	return metricService
}
