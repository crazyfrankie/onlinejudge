//go:build wireinject

package sms

import (
	"github.com/crazyfrankie/onlinejudge/internal/sms/service"
	"github.com/crazyfrankie/onlinejudge/internal/sms/service/failover"
	"github.com/crazyfrankie/onlinejudge/internal/sms/service/memory"
	"github.com/crazyfrankie/onlinejudge/internal/sms/service/metric"
	"github.com/crazyfrankie/onlinejudge/internal/sms/service/ratelimit"
	ratelimit2 "github.com/crazyfrankie/onlinejudge/pkg/ratelimit"
	"github.com/google/wire"
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

func NewModule(limiter ratelimit2.Limiter) *Module {
	wire.Build(
		InitSMS,

		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
