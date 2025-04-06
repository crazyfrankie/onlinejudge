package response

import (
	"net/http"
	"strconv"

	"github.com/crazyfrankie/gem/gerrors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/crazyfrankie/onlinejudge/common/constant"
)

type Response struct {
	Code    int32       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

var vector *prometheus.CounterVec

func InitCouter(opt prometheus.CounterOpts) {
	vector = prometheus.NewCounterVec(opt, []string{"code"})
	prometheus.MustRegister(vector)
}

func Success(ctx *gin.Context, data interface{}) {
	// 监控
	vector.WithLabelValues(strconv.Itoa(int(constant.Success.Code))).Inc()

	ctx.JSON(http.StatusOK, Response{
		Code:    constant.Success.Code,
		Message: constant.Success.Message,
		Data:    data,
	})
}

func Error(ctx *gin.Context, err error) {
	// 使用类型断言判断是否为业务错误
	if businessErr, ok := gerrors.FromBizStatusError(err); ok {
		ctx.JSON(http.StatusOK, Response{
			Code:    businessErr.BizStatusCode(),
			Message: businessErr.Error(),
		})

		// 监控
		vector.WithLabelValues(strconv.Itoa(int(businessErr.BizStatusCode()))).Inc()

		return
	}

	ctx.JSON(http.StatusOK, Response{
		Code:    0,
		Message: err.Error(),
	})
}
