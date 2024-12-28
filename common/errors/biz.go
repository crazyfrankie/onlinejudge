package errors

import (
	"github.com/crazyfrankie/gem/gerrors"

	"github.com/crazyfrankie/onlinejudge/common/constant"
)

type BizError struct {
	gerrors.BizErrorIface
}

func NewBizError(errCode constant.ErrorCode) *BizError {
	return &BizError{
		BizErrorIface: gerrors.NewBizError(errCode.Code, errCode.Message),
	}
}
