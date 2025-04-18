package ioc

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/crazyfrankie/onlinejudge/config"
)

func InitRedis() redis.Cmdable {
	client := redis.NewClient(&redis.Options{
		Addr:         config.GetConf().Redis.Address,
		Password:     "",
		MinIdleConns: config.GetConf().Redis.MinIdleConns,
		PoolSize:     config.GetConf().Redis.PoolSize,
		DialTimeout:  time.Minute * 5,
	})

	// tracing instrumentation
	if err := redisotel.InstrumentTracing(client); err != nil {
		panic(fmt.Sprintf("Failed to create Prometheus exporter: %v", err))
	}

	exporter, err := prometheus.New(prometheus.WithNamespace("cfc_studio_frank"))
	if err != nil {
		panic(err)
	}
	meterProvider := metric.NewMeterProvider(metric.WithReader(exporter))
	// metrics instrumentation.
	if err := redisotel.InstrumentMetrics(client, redisotel.WithMeterProvider(meterProvider)); err != nil {
		panic(fmt.Sprintf("Failed to instrument Redis metrics: %v", err))
	}

	return client
}
