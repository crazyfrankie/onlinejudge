package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/onlinejudge/internal/judgement/domain"
)

type SubmitCache interface {
	Set(ctx context.Context, userId uint64, hashKey string, evals []domain.RemoteEvaluation) error
	Get(ctx context.Context, userId uint64, hashKey string) ([]domain.RemoteEvaluation, error)
}

type SubmissionCache struct {
	cmd redis.Cmdable
}

func NewSubmitCache(cmd redis.Cmdable) SubmitCache {
	return &SubmissionCache{
		cmd: cmd,
	}
}

func (cache *SubmissionCache) Set(ctx context.Context, userId uint64, hashKey string, evals []domain.RemoteEvaluation) error {
	cacheKey := fmt.Sprintf("%d:%s", userId, hashKey)

	// 序列化
	data, err := sonic.Marshal(evals)
	if err != nil {
		return fmt.Errorf("failed to marshal evaluations: %w", err)
	}

	// 存入 Redis
	err = cache.cmd.Set(ctx, cacheKey, data, time.Minute).Err()
	return err
}

func (cache *SubmissionCache) Get(ctx context.Context, userId uint64, hashKey string) ([]domain.RemoteEvaluation, error) {
	cacheKey := fmt.Sprintf("%d:%s", userId, hashKey)

	data, err := cache.cmd.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, err
		}
		// 处理其他错误
		return nil, err
	}

	// 反序列化
	var evals []domain.RemoteEvaluation
	err = sonic.Unmarshal([]byte(data), &evals)
	if err != nil {
		// 处理反序列化错误
		return nil, err
	}

	return evals, err
}
