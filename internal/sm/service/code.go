package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"

	"github.com/crazyfrankie/onlinejudge/common/constant"
	er "github.com/crazyfrankie/onlinejudge/common/errors"
	"github.com/crazyfrankie/onlinejudge/internal/sm/repository"
	svc "github.com/crazyfrankie/onlinejudge/internal/sm/service/sms/service"
)

const (
	codeTplId = ""
	secretKey = "BgrTwHrRffd6LMXZWXGJCaKZHGb5p5h8"
)

var (
	ErrSendTooMany   = repository.ErrSendTooMany
	ErrVerifyTooMany = repository.ErrVerifyTooMany
)

type CodeService interface {
	Send(ctx context.Context, biz, receiver string) error
	Verify(ctx context.Context, biz, phone, inputCode string) error
	GenerateCode() string
	GenerateHMAC(code, key string) string
}

type CodeSvc struct {
	repo repository.CodeRepository
	sms  svc.Service
}

func NewCodeService(r repository.CodeRepository, sms svc.Service) CodeService {
	return &CodeSvc{
		repo: r,
		sms:  sms,
	}
}

func (svc *CodeSvc) Send(ctx context.Context, biz, receiver string) error {
	ctx, span := otel.Tracer("onlinejudge/user/service").Start(ctx, "CodeService/Verify")
	defer span.End()

	code := svc.GenerateCode()
	enCode := svc.GenerateHMAC(code, secretKey)

	err := svc.repo.Store(ctx, biz, receiver, enCode)
	if err != nil {
		if errors.Is(err, ErrSendTooMany) {
			return er.NewBizError(constant.ErrUserForbidden)
		}
		return er.NewBizError(constant.ErrCodeInternalServer)
	}

	err = svc.sms.Send(ctx, codeTplId, []string{code}, receiver)
	if err != nil {
		return er.NewBizError(constant.ErrCodeInternalServer)
	}

	return nil
}

func (svc *CodeSvc) Verify(ctx context.Context, biz, phone, inputCode string) error {
	ctx, span := otel.Tracer("onlinejudge/user/service").Start(ctx, "CodeService/Verify")
	defer span.End()

	// 拿到 code 后加密再去跟 redis 中的 code 进行对比
	enCode := svc.GenerateHMAC(inputCode, secretKey)

	_, err := svc.repo.Verify(ctx, biz, phone, enCode)
	if err != nil {
		if errors.Is(err, ErrVerifyTooMany) {
			return er.NewBizError(constant.ErrVerifyTooMany)
		} else {
			return er.NewBizError(constant.ErrCodeInternalServer)
		}
	}

	return nil
}

func (svc *CodeSvc) GenerateCode() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	var code strings.Builder
	for i := 0; i < 6; i++ {
		digit := rand.Intn(10)
		code.WriteString(strconv.Itoa(digit))
	}
	return code.String()
}

func (svc *CodeSvc) GenerateHMAC(code, key string) string {
	// 创建一个新的 HMAC 哈希对象，使用 SHA-256 哈希算法，并以 key 作为密钥。
	h := hmac.New(sha256.New, []byte(key))

	// 将输入的 code 数据写入到 HMAC 哈希对象中，进行哈希计算。
	h.Write([]byte(code))

	// 计算哈希值并返回其十六进制表示形式。
	return hex.EncodeToString(h.Sum(nil))
}
