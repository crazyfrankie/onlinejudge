package biz

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"oj/user/repository"
	smsSvc "oj/user/service/sms"
)

const codeTplId = ""

type CodeService struct {
	repo *repository.CodeRepository
	sms  smsSvc.Service
}

func NewCodeService(r *repository.CodeRepository) *CodeService {
	return &CodeService{
		repo: r,
	}
}

func (svc *CodeService) Send(ctx context.Context, biz, phone, code string) error {
	// 生成一个验证码
	svc.generateCode()

	// 塞进去 Redis
	err := svc.repo.Store(ctx, biz, phone, code)
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

	return fmt.Sprintf("%6d", randomNumber)
}
