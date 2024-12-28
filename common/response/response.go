package response

import (
	"net/http"

	"github.com/crazyfrankie/gem/gerrors"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/onlinejudge/common/constant"
)

type Response struct {
	Code    int32       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(ctx *gin.Context, data interface{}) {
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
		return
	}
}
