package repository

import (
	"context"
	"github.com/crazyfrankie/onlinejudge/internal/judgement/domain"
	"github.com/crazyfrankie/onlinejudge/internal/judgement/repository/cache"
)

type SubmitRepository interface {
	StoreEvaluation(ctx context.Context, userId uint64, code string, evals []domain.Evaluation) error
	AcquireEvaluation(ctx context.Context, userId uint64, hashKey string) ([]domain.Evaluation, error)
}

type SubmissionRepo struct {
	cache cache.SubmitCache
}

func NewSubmitRepository(cache cache.SubmitCache) SubmitRepository {
	return &SubmissionRepo{
		cache: cache,
	}
}

func (repo *SubmissionRepo) StoreEvaluation(ctx context.Context, userId uint64, hashKey string, evals []domain.Evaluation) error {
	return repo.cache.Set(ctx, userId, hashKey, evals)
}

func (repo *SubmissionRepo) AcquireEvaluation(ctx context.Context, userId uint64, hashKey string) ([]domain.Evaluation, error) {
	return repo.cache.Get(ctx, userId, hashKey)
}
