// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package sms

import (
	"github.com/crazyfrankie/onlinejudge/internal/sms/service"
	"github.com/crazyfrankie/onlinejudge/internal/sms/service/failover"
	"github.com/crazyfrankie/onlinejudge/internal/sms/service/memory"
	"github.com/crazyfrankie/onlinejudge/internal/sms/service/metric"
	ratelimit2 "github.com/crazyfrankie/onlinejudge/internal/sms/service/ratelimit"
	"github.com/crazyfrankie/onlinejudge/pkg/ratelimit"
)

// Injectors from wire.go:

func NewModule(limiter ratelimit.Limiter) *Module {
	service := InitSMS(limiter)
	module := &Module{
		SmsSvc: service,
	}
	return module
}

// wire.go:

func InitSMS(limiter ratelimit.Limiter) service.Service {
	memoryService := memory.NewService()

	services := []service.Service{
		memoryService,
	}
	rateLimitService := ratelimit2.NewService(memoryService, limiter)
	failOverService := failover.NewFailOver(append(services, rateLimitService))
	metricService := metric.NewProemtheusDecortor(failOverService)

	return metricService
}