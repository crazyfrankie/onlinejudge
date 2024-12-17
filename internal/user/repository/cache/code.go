package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	ca "github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	ErrSendTooMany   = errors.New("send too frequency")
	ErrVerifyTooMany = errors.New("too many verifications")
	ErrKeyNotFound   = errors.New("key expired or not found")
)

//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
	key(biz, phone string) string
}

// RedisCodeCache 基于 Redis 实现
type RedisCodeCache struct {
	client redis.Cmdable
}

func NewRedisCodeCache(client redis.Cmdable) CodeCache {
	return &RedisCodeCache{
		client: client,
	}
}

func (c *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := c.client.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0:
		// 毫无问题
		return nil
	case -1:
		// 发送太频繁
		return ErrSendTooMany
	default:
		return errors.New("system errors")
	}
}

func (c *RedisCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	ok, err := c.client.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, inputCode).Int()
	if err != nil {
		return false, err
	}
	switch ok {
	case 0:
		// 没问题
		return true, nil
	case -1:
		// 如果频繁出这个错误代表有人搞你 需要告警
		return false, ErrVerifyTooMany
	}
	return false, errors.New("system errors")
}

func (c *RedisCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

// MemCodeCache 基于本地内存的实现
type MemCodeCache struct {
	cache *ca.Cache
}

type CodeVal struct {
	Code      string
	createdAt int64
}

func NewMemCodeCache(ce *ca.Cache) CodeCache {
	return &MemCodeCache{cache: ce}
}

func (cm *MemCodeCache) Set(ctx context.Context, biz, phone, code string) error {
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

func (cm *MemCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	val, ok := cm.cache.Get(cm.key(biz, phone))
	if !ok {
		return false, ErrKeyNotFound
	}
	if val.(CodeVal).Code == inputCode {
		return true, nil
	}
	return false, errors.New("system errors")
}

func (cm *MemCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
