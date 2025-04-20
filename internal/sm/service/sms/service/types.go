package service

import (
	"context"
)

type Service interface {
	Send(ctx context.Context, biz string, args []string, numbers ...string) error
}
