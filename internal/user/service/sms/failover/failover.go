/*
短信服务商切换策略
FailOver
*/

package failover

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/crazyfrankie/onlinejudge/internal/user/service/sms"
)

type SMSFailOver struct {
	svc []sms.Service
	idx uint64
}

func NewFailOver(svc []sms.Service) sms.Service {
	return &SMSFailOver{
		svc: svc,
	}
}

func (f *SMSFailOver) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 原子操作，并发安全
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svc))
	for i := idx; i < length+idx; i++ {
		svc := f.svc[int(i%length)]
		err := svc.Send(ctx, tplId, args, numbers...)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			// 调用者的超时时间到了
			// 调用者主动取消了
			return err
		}
	}
	return errors.New("all sms failed")
}
