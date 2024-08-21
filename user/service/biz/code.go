package biz

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"oj/user/repository"
	smsSvc "oj/user/service/sms"
)

const codeTplId = ""

var (
	ErrSendTooMany = repository.ErrSendTooMany
)

type CodeService struct {
	repo *repository.CodeRepository
	sms  smsSvc.Service
}

func NewCodeService(r *repository.CodeRepository, sms smsSvc.Service) *CodeService {
	return &CodeService{
		repo: r,
		sms:  sms,
	}
}

func (svc *CodeService) Send(ctx context.Context, biz, phone string) error {
	// 生成一个验证码
	code := svc.generateCode()

	// 塞进去 Redis
	err := svc.repo.Store(ctx, biz, phone, code)
	if errors.Is(err, ErrSendTooMany) {
		return ErrSendTooMany
	}
	if err != nil {
		return err
	}

	// 发送出去
	err = svc.sms.Send(ctx, codeTplId, []string{code}, phone)

	return err
}

func (svc *CodeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *CodeService) generateCode() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	randomNumber := rand.Intn(1000000)

	return fmt.Sprintf("%06d", randomNumber)
}
