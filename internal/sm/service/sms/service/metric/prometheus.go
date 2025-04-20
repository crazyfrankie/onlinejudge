package metric

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/crazyfrankie/onlinejudge/internal/sm/service/sms/service"
)

type PrometheusDecortor struct {
	svc    service.Service
	vector *prometheus.SummaryVec
}

func NewProemtheusDecortor(svc service.Service) *PrometheusDecortor {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "cfc_studio_frank",
		Subsystem: "onlinejudge",
		Name:      "sms_resp_time",
		Help:      "统计 SMS 服务的性能数据",
	}, []string{"biz"})
	prometheus.MustRegister(vector)
	return &PrometheusDecortor{
		svc:    svc,
		vector: vector,
	}
}

func (p *PrometheusDecortor) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Milliseconds()
		p.vector.WithLabelValues(biz).Observe(float64(duration))
	}()
	return p.svc.Send(ctx, biz, args, numbers...)
}
