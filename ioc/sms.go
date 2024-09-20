package ioc

import (
	"oj/internal/user/service/sms"
	"oj/internal/user/service/sms/failover"
	"oj/internal/user/service/sms/memory"
	"oj/internal/user/service/sms/ratelimit"
	ratelimit2 "oj/pkg/ratelimit"
)

func InitSMSService(limiter ratelimit2.Limiter) sms.Service {
	memoryService := memory.NewService()

	services := []sms.Service{
		memoryService,
	}
	rateLimitService := ratelimit.NewService(memoryService, limiter)
	failOverService := failover.NewFailOver(append(services, rateLimitService))
	return failOverService
}
