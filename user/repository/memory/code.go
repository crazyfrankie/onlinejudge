package memory

import (
	"context"
	"errors"
	"fmt"
	"github.com/patrickmn/go-cache"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key expired or not found")
	ErrSendTooMany = errors.New("send too many")
)

type CodeMem struct {
	cache *cache.Cache
}

type CodeVal struct {
	Code      string
	createdAt int64
}

func NewCodeMem(ce *cache.Cache) *CodeMem {
	return &CodeMem{cache: ce}
}

func (cm *CodeMem) Set(ctx context.Context, biz, phone, code string) error {
	key := cm.key(biz, phone)

	// 检查是否存在
	if existingVal, found := cm.cache.Get(key); found {
		val := existingVal.(CodeVal)
		elapsed := time.Now().Unix() - val.createdAt

		if elapsed < 60 {
			return ErrSendTooMany
		}
	}

	cm.cache.Set(key, CodeVal{Code: code, createdAt: time.Now().Unix()}, time.Minute*10)
	return nil
}

func (cm *CodeMem) Verify(ctx context.Context, biz, phone, inputCode string) (string, error) {
	val, ok := cm.cache.Get(cm.key(biz, phone))
	if !ok {
		return "", ErrKeyNotFound
	}
	if val.(CodeVal).Code == inputCode {
		return val.(CodeVal).Code, nil
	}
	return "", errors.New("system error")
}

func (cm *CodeMem) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
