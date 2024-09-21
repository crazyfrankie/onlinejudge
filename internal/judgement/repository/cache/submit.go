package cache

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"oj/internal/judgement/domain"
)

type SubmitCache interface {
	Set(ctx context.Context, eval domain.Evaluation) error
	Get(ctx context.Context, id uint64) (domain.Evaluation, error)
}

type SubmissionCache struct {
	cmd redis.Cmdable
}

func NewSubmitCache(cmd redis.Cmdable) SubmitCache {
	return &SubmissionCache{
		cmd: cmd,
	}
}

func (cache *SubmissionCache) Set(ctx context.Context, eval domain.Evaluation) error {
	return errors.New("error")
}

func (cache *SubmissionCache) Get(ctx context.Context, id uint64) (domain.Evaluation, error) {
	return domain.Evaluation{}, errors.New("error")
}
