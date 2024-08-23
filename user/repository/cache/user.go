package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"oj/user/domain"
)

type UserCache struct {
	client redis.Cmdable
}

func NewUserCache(client redis.Cmdable) *UserCache {
	return &UserCache{
		client: client,
	}
}

func (cache *UserCache) Get(ctx context.Context, id uint64) (domain.User, error) {
	key := cache.key(id)
	val, err := cache.client.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}

	var user domain.User
	err = json.Unmarshal([]byte(val), &user)
	return user, err
}

func (cache *UserCache) Set(ctx context.Context, user domain.User) error {
	val, err := json.Marshal(user)
	if err != nil {
		return err
	}

	key := cache.key(user.Id)

	return cache.client.Set(ctx, key, val, time.Minute*10).Err()
}

func (cache *UserCache) key(id uint64) string {
	return fmt.Sprintf("user:info:%d", id)
}
