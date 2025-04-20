package ioc

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
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

	exporter, err := otlpmetricgrpc.New(context.Background(),
		otlpmetricgrpc.WithEndpoint("otel-collector:4317"),
		otlpmetricgrpc.WithInsecure())
	if err != nil {
		panic(err)
	}
 
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter,
			metric.WithTimeout(time.Second*10),
			metric.WithInterval(time.Second*30))),
	)
	// metrics instrumentation.
	if err := redisotel.InstrumentMetrics(client, redisotel.WithMeterProvider(meterProvider)); err != nil {
		panic(fmt.Sprintf("Failed to instrument Redis metrics: %v", err))
	}

	return client
}
