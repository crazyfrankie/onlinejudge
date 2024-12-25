// Package ratelimit
/*
装饰器模式实现
短信在客户端的限流
*/
package ratelimit

import (
	"context"
	"fmt"

	"github.com/crazyfrankie/onlinejudge/internal/user/service/sms"
	"github.com/crazyfrankie/onlinejudge/pkg/ratelimit"
)

type Service struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &Service{
		svc:     svc,
		limiter: limiter,
	}
}

func (svc *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 可在这增加新特性
	limiter, err := svc.limiter.Limit(ctx, "sms:tencent")
	if err != nil {
		// 系统错误
		// 可以限流：保守策略，下游比较坑时
		// 可以不限：下游很强，业务可用性很高时，尽量容错策略
		return fmt.Errorf("短信服务判断是否异常 %w", err)
	}
	if limiter {
		return fmt.Errorf("短信服务触发限流 %w", err)
	}

	err = svc.svc.Send(ctx, tplId, args, numbers...)
	// 也可在这增加新特性
	return err
}
