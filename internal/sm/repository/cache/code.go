package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	ca "github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
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
	key := c.key(biz, phone)

	// 检查是否在60秒内重复发送
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	// 如果key存在且剩余时间大于540秒(600-60)，说明是60秒内重复发送
	if ttl > 540 {
		return ErrSendTooMany
	}

	// 使用管道执行多个命令
	pipe := c.client.Pipeline()
	// 设置验证码，10分钟过期
	pipe.Set(ctx, key, code, time.Minute*10)
	// 设置验证次数为3次，同样10分钟过期
	pipe.Set(ctx, key+":cnt", 3, time.Minute*10)

	_, err = pipe.Exec(ctx)
	return err
}

func (c *RedisCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	key := c.key(biz, phone)
	cntKey := key + ":cnt"

	// 使用管道获取验证码和验证次数
	pipe := c.client.Pipeline()
	codeCmd := pipe.Get(ctx, key)
	cntCmd := pipe.Get(ctx, cntKey)
	_, err := pipe.Exec(ctx)

	// 处理key不存在的情况
	if err == redis.Nil {
		return false, ErrKeyNotFound
	}
	if err != nil {
		return false, err
	}

	// 获取验证次数
	cnt, err := cntCmd.Int()
	if err != nil {
		return false, err
	}
	if cnt < 0 {
		return false, ErrVerifyTooMany
	}

	// 获取验证码
	code, err := codeCmd.Result()
	if err != nil {
		return false, err
	}

	// 验证码匹配
	if code == inputCode {
		// 验证成功，将计数器设为-1
		err = c.client.Set(ctx, cntKey, -1, time.Minute*10).Err()
		return err == nil, err
	}

	// 验证码不匹配，次数减1
	err = c.client.Decr(ctx, cntKey).Err()
	return false, err
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
