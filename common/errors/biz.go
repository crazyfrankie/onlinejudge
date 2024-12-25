package errors

import (
	"errors"

	"github.com/crazyfrankie/onlinejudge/common/constant"
)

type BusinessError struct {
	ErrorCode constant.ErrorCode
}

// Error 实现 errors 接口
func (e *BusinessError) Error() string {
	return e.ErrorCode.Message
}

// Code 用于类型断言
func (e *BusinessError) Code() int {
	return e.ErrorCode.Code
}

// NewBusinessError 创建业务错误的便捷方法
func NewBusinessError(errCode constant.ErrorCode) *BusinessError {
	return &BusinessError{ErrorCode: errCode}
}

// IsBusinessError 一些常用的错误判断方法
func IsBusinessError(err error) (*BusinessError, bool) {
	if err == nil {
		return nil, false
	}
	var be *BusinessError
	ok := errors.As(err, &be)
	return be, ok
}
