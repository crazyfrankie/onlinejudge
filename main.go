package main

import (
	"context"
	"log"
	"net/http"
	"syscall"
	"time"

	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/crazyfrankie/onlinejudge/config"
	"github.com/crazyfrankie/onlinejudge/ioc"
)

func main() {
	g := &run.Group{}

	closeFunc := ioc.InitOTEL()
	app := ioc.InitApp()

	g.Add(func() error {
		http.Handle("/metrics", promhttp.Handler())
		return http.ListenAndServe("0.0.0.0"+config.GetConf().Server.Metrics, nil)
	}, func(err error) {
		// Prometheus 服务器通常不需要特殊关闭处理
	})

	server := &http.Server{
		Addr:    "0.0.0.0" + config.GetConf().Server.Addr,
		Handler: app.Server,
	}
	g.Add(func() error {
		log.Printf("Server is running at http://localhost%s\n", config.GetConf().Server.Addr)
		return server.ListenAndServe()
	}, func(err error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("failed to shutdown main server: %v", err)
		}
	})

	// start consumers
	for _, consumer := range app.Consumers {
		err := consumer.Start()
		if err != nil {
			panic(err)
		}
	}

	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	// 运行所有服务
	if err := g.Run(); err != nil {
		log.Printf("program interrupted, err:%s", err)
	}

	// 关闭 OTEL 连接
	newCtx, cancelFunc := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFunc()
	closeFunc(newCtx)

	log.Println("Server exited gracefully")
}
