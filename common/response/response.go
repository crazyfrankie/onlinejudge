package response

import (
	"net/http"

	"github.com/crazyfrankie/onlinejudge/common/constant"
	"github.com/crazyfrankie/onlinejudge/common/errors"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
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
	if businessErr, ok := errors.IsBusinessError(err); ok {
		ctx.JSON(http.StatusOK, Response{
			Code:    businessErr.Code(),
			Message: businessErr.Error(),
		})
		return
	}

	// 非业务错误统一返回服务器错误
	ctx.JSON(http.StatusOK, Response{
		Code:    constant.ErrInternalServer.Code,
		Message: constant.ErrInternalServer.Message,
	})
}
