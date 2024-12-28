package web

import (
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"time"
)

type MetricHandler struct {
}

func (m *MetricHandler) RegisterRoute(r *gin.Engine) {
	testGroup := r.Group("api/test")
	{
		testGroup.GET("metric", m.Metric())
	}
}

func (m *MetricHandler) Metric() gin.HandlerFunc {
	return func(c *gin.Context) {
		sleep := rand.Int31n(1000)
		time.Sleep(time.Millisecond * time.Duration(sleep))
		c.String(http.StatusOK, "OK")
	}
}
