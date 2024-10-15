package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"oj/internal/judgement/domain"
	"time"
)

type LocalSubmitCache interface {
	Set(ctx context.Context, userId, problemId uint64, evals []domain.Evaluation) error
	Get(ctx context.Context, userId, problemId uint64) ([]domain.Evaluation, error)

	key(userId, problemId uint64) string
}

type LocalSubmissionCache struct {
	cmd redis.Cmdable
}

func NewLocalSubmitCache(cmd redis.Cmdable) LocalSubmitCache {
	return &LocalSubmissionCache{
		cmd: cmd,
	}
}

func (cache *LocalSubmissionCache) Set(ctx context.Context, userId, problemId uint64, evals []domain.Evaluation) error {
	key := cache.key(userId, problemId)

	val, err := json.Marshal(evals)
	if err != nil {
		return err
	}

	err = cache.cmd.Set(ctx, key, val, time.Minute*10).Err()
	return err
}

func (cache *LocalSubmissionCache) Get(ctx context.Context, userId, problemId uint64) ([]domain.Evaluation, error) {
	var evals []domain.Evaluation

	key := cache.key(userId, problemId)

	val, err := cache.cmd.Get(ctx, key).Result()
	if err != nil {
		return evals, err
	}

	err = json.Unmarshal([]byte(val), &evals)
	return evals, err
}

func (cache *LocalSubmissionCache) key(userId, problemId uint64) string {
	return fmt.Sprintf("user:%d:problem:%d:code", userId, problemId)
}
