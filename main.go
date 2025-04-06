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

	"github.com/crazyfrankie/onlinejudge/ioc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
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
	log.Printf("Server is running address:%s", "http://localhost:8082")

	// 创建通道监听信号
	quit := make(chan os.Signal, 1)

	// 监听信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞直到收到信号
	<-quit
	log.Println("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 优雅地关闭服务器
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced shutting down:%s", err)
	}

	// 关闭 OTEL 连接
	newCtx, cancelFunc := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFunc()
	closeFunc(newCtx)

	log.Println("Server exited gracefully")
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe("0.0.0.0:8081", nil)
		if err != nil {
			panic(err)
		}
	}()
}
