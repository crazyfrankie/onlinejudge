package service

import (
	"context"
)

type InteractiveService struct {
}

func NewInteractiveService() *InteractiveService {
	return &InteractiveService{}
}

func (svc *InteractiveService) IncrReadCnt(ctx context.Context, biz string, bizId int) error {
	panic("")
}
