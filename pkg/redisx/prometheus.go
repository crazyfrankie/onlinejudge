/*
使用 Prometheus 进行 Redis 执行时间、缓存命中率的监控
*/

package redisx

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type PrometheusHook struct {
	vector *prometheus.SummaryVec
}

func NewPrometheusHook(opt prometheus.SummaryOpts) *PrometheusHook {
	// key_exist 是否命中缓存
	vector := prometheus.NewSummaryVec(opt, []string{"cmd", "key_exist"})
	prometheus.MustRegister(vector)
	return &PrometheusHook{
		vector: vector,
	}
}

func (p *PrometheusHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (p *PrometheusHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		// 在 Redis 执行之前
		startTime := time.Now()

		var err error
		defer func() {
			duration := time.Since(startTime).Microseconds()
			// 是否命中缓存
			if cmd.Name() == "get" {
				keyExists := errors.Is(err, redis.Nil)
				p.vector.WithLabelValues(cmd.Name(), strconv.FormatBool(keyExists)).Observe(float64(duration))
			} else {
				p.vector.WithLabelValues(cmd.Name()).Observe(float64(duration))
			}
		}()
		// 在 Redis 执行之后
		err = next(ctx, cmd)

		return err
	}
}

func (p *PrometheusHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}
}
