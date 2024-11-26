package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"oj/internal/user/repository"
	smsSvc "oj/internal/user/service/sms"
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
	Send(ctx context.Context, biz, receiver string) *CodeServiceError
	Verify(ctx context.Context, biz, phone, inputCode string) *CodeServiceError
	GenerateCode() string
	GenerateHMAC(code, key string) string
}

type CodeSvc struct {
	repo repository.CodeRepository
	sms  smsSvc.Service
}

type CodeServiceError struct {
	Code    int
	Message string
}

func NewCodeService(r repository.CodeRepository, sms smsSvc.Service) CodeService {
	return &CodeSvc{
		repo: r,
		sms:  sms,
	}
}

func (svc *CodeSvc) Send(ctx context.Context, biz, receiver string) *CodeServiceError {
	// 生成一个验证码
	code := svc.GenerateCode()

	// 加密
	enCode := svc.GenerateHMAC(code, secretKey)

	// 塞进去 Redis
	err := svc.repo.Store(ctx, biz, receiver, enCode)
	if err != nil {
		if errors.Is(err, ErrSendTooMany) {
			return &CodeServiceError{Code: http.StatusTooManyRequests, Message: "send too many"}
		} else {
			return &CodeServiceError{Code: http.StatusBadRequest, Message: "system error"}
		}
	}

	// 发送出去
	err = svc.sms.Send(ctx, codeTplId, []string{code}, receiver)
	if err != nil {
		return &CodeServiceError{Code: http.StatusInternalServerError, Message: "failed to send"}
	}

	return nil
}

func (svc *CodeSvc) Verify(ctx context.Context, biz, phone, inputCode string) *CodeServiceError {
	// 拿到 code 后加密再去跟 redis 中的 code 进行对比
	enCode := svc.GenerateHMAC(inputCode, secretKey)

	_, err := svc.repo.Verify(ctx, biz, phone, enCode)
	if err != nil {
		if errors.Is(err, ErrVerifyTooMany) {
			return &CodeServiceError{Code: http.StatusTooManyRequests, Message: "verify too many"}
		} else {
			return &CodeServiceError{Code: http.StatusBadRequest, Message: "system error"}
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
