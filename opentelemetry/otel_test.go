package opentelemetry

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	res, err := newRource("demo", "v0.0.1")
	require.NoError(t, err)

	propagator := newPropagator()
	otel.SetTextMapPropagator(propagator)

	traceProvider, er := newTraceProvider(res)
	require.NoError(t, er)
	otel.SetTracerProvider(traceProvider)
	defer func(traceProvider *trace.TracerProvider, ctx context.Context) {
		err := traceProvider.Shutdown(ctx)
		if err != nil {
			log.Printf("%s", err.Error())
		}
	}(traceProvider, context.Background())

	server := gin.Default()

	server.Use(otelgin.Middleware("demo"))

	server.GET("/test", func(c *gin.Context) {
		// 创建一个 Trace span 来捕获这个请求
		//_, span := otel.Tracer("demo").Start(c.Request.Context(), "GET /test")
		//defer span.End()

		fmt.Println("a")
		time.Sleep(time.Second)
		fmt.Println("b")
		c.JSON(http.StatusOK, "OK")
	})

	// 不能使用 Gin 的 Run
	// 因为使用了 otelhttp.NewHandler 自动管理 Span 的创建和上下文的替换, 而 otelhttp.NewHandler 返回的是它封装的 http.Handler
	// 如果想要在所有路由中使用 otelhttp 提供的自动管理功能,必须使用它的 Handler 作为 HTTP 服务器的 Handler
	// 而 Gin 的 Run 使用的 Handler 是不完全兼容的 http.Handler
	server.Run(":9098")
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
		trace.WithBatcher(exporter, trace.WithBatchTimeout(time.Second)), trace.WithResource(res))

	return traceProvider, nil
}
