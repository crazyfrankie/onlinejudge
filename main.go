package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/crazyfrankie/onlinejudge/ioc"
)

func main() {
	err := godotenv.Load("config/.env")
	if err != nil {
		panic(err)
	}

	initLogger()
	initPrometheus()

	closeFunc := ioc.InitOTEL()

	app := ioc.InitApp()

	server := &http.Server{
		Addr:    "0.0.0.0:8082",
		Handler: app.Server,
	}

	// start consumers
	for _, consumer := range app.Consumers {
		err := consumer.Start()
		if err != nil {
			panic(err)
		}
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	zap.L().Info("Server is running", zap.String("address", "http://localhost:8082"))

	// 创建通道监听信号
	quit := make(chan os.Signal, 1)

	// 监听信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞直到收到信号
	<-quit
	zap.L().Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 优雅地关闭服务器
	if err := server.Shutdown(ctx); err != nil {
		zap.L().Error("Server forced shutting down", zap.Error(err))
	}

	// 关闭 OTEL 连接
	newCtx, cancelFunc := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFunc()
	closeFunc(newCtx)

	zap.L().Info("Server exited gracefully")
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":8081", nil)
		if err != nil {
			panic(err)
		}
	}()
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}
