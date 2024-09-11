package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"oj/internal/problem/domain"
)

type ProblemCache interface {
	Set(ctx context.Context, problem domain.Problem) error
	Get(ctx context.Context, id uint64) (domain.Problem, error)
	key(id uint64) string
}

type ProblemCe struct {
	cmd redis.Cmdable
}

func NewProblemCache(cmd redis.Cmdable) ProblemCache {
	return &ProblemCe{
		cmd: cmd,
	}
}

func (cache *ProblemCe) Set(ctx context.Context, problem domain.Problem) error {
	key := cache.key(problem.Id)

	val, err := json.Marshal(problem)
	if err != nil {
		return err
	}

	return cache.cmd.Set(ctx, key, val, time.Minute*15).Err()
}

func (cache *ProblemCe) Get(ctx context.Context, id uint64) (domain.Problem, error) {
	key := cache.key(id)
	val, err := cache.cmd.Get(ctx, key).Result()
	if err != nil {
		return domain.Problem{}, err
	}

	var pm domain.Problem
	err = json.Unmarshal([]byte(val), &pm)
	return pm, err
}

func (cache *ProblemCe) key(id uint64) string {
	return fmt.Sprintf("problem:info:%d", id)
}
