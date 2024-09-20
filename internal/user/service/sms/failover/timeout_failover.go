package failover

import (
	"context"
	"errors"
	"sync/atomic"

	"oj/internal/user/service/sms"
)

type TimeOutService struct {
	svc []sms.Service
	idx int32
	// 计数器
	cnt int32

	// 阈值
	threshold int32
}

func NewTimeOutService(svc []sms.Service) sms.Service {
	return &TimeOutService{
		svc:       svc,
		threshold: 100,
	}
}

func (t *TimeOutService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)

	if cnt > t.threshold {
		// 切换下标，往后挪了一个
		newIdx := (idx + 1) % int32(len(t.svc))
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			// 成功往后挪了一位
			atomic.StoreInt32(&t.cnt, 0)
		}
		// 出现并发了，别人换了
		idx = newIdx
		// idx = atomic.LoadInt32(&t.idx)
	}

	svc := t.svc[idx]
	err := svc.Send(ctx, tplId, args, numbers...)
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		atomic.AddInt32(&t.cnt, 1)
		return nil
	case err == nil:
		// 连续状态被打断了
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	default:
		// 不知道什么错误
		// 有些情况可以考虑换下一个
		// - 超时，可能是偶发的，我尽量再试试
		// - 非超时，我直接换下一个
		return err
	}
}
