package response

import (
	"github.com/crazyfrankie/onlinejudge/pkg/zapx"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"sync"

	"github.com/crazyfrankie/gem/gerrors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/crazyfrankie/onlinejudge/common/constant"
)

var (
	logger *zapx.Logger
	once   sync.Once
)

func init() {
	once.Do(func() {
		logger = zapx.NewLogger(zap.NewProductionConfig())
	})
}

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

func ErrorWithLog(ctx *gin.Context, name string, msg string, err error) {
	logger.Error(ctx.Request.Context(), name, msg, zap.Error(err))

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

func SuccessWithLog(ctx *gin.Context, data any, name string, msg string, fields ...zap.Field) {
	logger.Info(ctx.Request.Context(), name, msg, fields...)

	vector.WithLabelValues(strconv.Itoa(int(constant.Success.Code))).Inc()

	ctx.JSON(http.StatusOK, Response{
		Code:    constant.Success.Code,
		Message: constant.Success.Message,
		Data:    data,
	})
}
