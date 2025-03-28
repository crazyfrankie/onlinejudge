package cache

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/onlinejudge/internal/user/domain"
)

type UserCache interface {
	Get(ctx context.Context, id uint64) (domain.User, error)
	Set(ctx context.Context, user domain.User) error
	Del(ctx context.Context, id uint64) error
	SetCheckState(ctx context.Context, phone string) error
	GetCheckState(ctx context.Context, phone string) (bool, error)
	key(id uint64) string
}

type RedisUserCache struct {
	client redis.Cmdable
}

func NewUserCache(client redis.Cmdable) UserCache {
	return &RedisUserCache{
		client: client,
	}
}

func (cache *RedisUserCache) Get(ctx context.Context, id uint64) (domain.User, error) {
	key := cache.key(id)
	val, err := cache.client.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}

	var user domain.User
	err = sonic.Unmarshal([]byte(val), &user)
	return user, err
}

func (cache *RedisUserCache) Set(ctx context.Context, user domain.User) error {
	val, err := sonic.Marshal(user)
	if err != nil {
		return err
	}

	key := cache.key(user.Id)

	return cache.client.Set(ctx, key, val, time.Minute*10).Err()
}

func (cache *RedisUserCache) Del(ctx context.Context, id uint64) error {
	key := cache.key(id)

	err := cache.client.Del(ctx, key).Err()

	return err
}

func (cache *RedisUserCache) key(id uint64) string {
	return fmt.Sprintf("user:info:%d", id)
}

func (cache *RedisUserCache) SetCheckState(ctx context.Context, phone string) error {
	key := fmt.Sprintf("verification_required:%s", phone)

	err := cache.client.Set(ctx, key, true, time.Minute*1).Err()
	if err != nil {
		return err
	}

	return nil
}

func (cache *RedisUserCache) GetCheckState(ctx context.Context, phone string) (bool, error) {
	key := fmt.Sprintf("verification_required:%s", phone)

	result, err := cache.client.Get(ctx, key).Bool()
	if err != nil {
		return false, err
	}

	return result, nil
}
