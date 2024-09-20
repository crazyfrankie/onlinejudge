package cache

import (
	"context"
	"oj/internal/judgement/domain"
)

type SubmitCache interface {
	Set(ctx context.Context, eval domain.Evaluation) error
	Get(ctx context.Context, id uint64) (domain.Evaluation, error)
}

type SubmissionCache struct {
}

func (cache *SubmissionCache) Set(ctx context.Context, eval domain.Evaluation) error {

}

func (cache *SubmissionCache) Get(ctx context.Context, id uint64) (domain.Evaluation, error) {

}
