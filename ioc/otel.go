package ioc

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"time"
)

func InitOTEL() func(ctx context.Context) {
	res, err := newRource("onlinejudge", "v0.0.1")
	if err != nil {
		panic(err)
	}
	propagator := newPropagator()
	otel.SetTextMapPropagator(propagator)

	traceProvider, er := newTraceProvider(res)
	if er != nil {
		panic(er)
	}
	otel.SetTracerProvider(traceProvider)

	return func(ctx context.Context) {
		err := traceProvider.Shutdown(ctx)
		if err != nil {
			return
		}
	}
}

func newRource(servicename, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceNameKey.String(servicename),
			semconv.ServiceVersionKey.String(serviceVersion)))
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{})
}

func newTraceProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	exporter, err := zipkin.New("http://localhost:9411/api/v2/spans")
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter, trace.WithBatchTimeout(time.Second*5)), trace.WithResource(res))

	return traceProvider, nil
}
