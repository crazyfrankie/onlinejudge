package errors

import (
	"errors"
	"github.com/crazyfrankie/onlinejudge/common/constant"
)

type BizError struct {
	errCode constant.ErrorCode
}

func (b *BizError) Error() string {
	return b.errCode.Message
}

func (b *BizError) StatusCode() int {
	return b.errCode.Status
}

func (b *BizError) BizCode() int32 {
	return b.errCode.Code
}

func FromBizStatusError(err error) (bizErr *BizError, ok bool) {
	if err == nil {
		return
	}
	ok = errors.As(err, &bizErr)
	return
}

func NewBizError(errCode constant.ErrorCode) *BizError {
	return &BizError{
		errCode: errCode,
	}
}
