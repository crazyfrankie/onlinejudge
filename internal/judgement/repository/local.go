package repository

import (
	"context"
	"oj/internal/judgement/domain"
	"oj/internal/judgement/repository/cache"
)

type LocalSubmitRepo interface {
	StoreEvaluationResult(ctx context.Context, userId, problemId uint64, evals []domain.Evaluation) error
	AcquireEvaluationResult(ctx context.Context, userId, problemId uint64) ([]domain.Evaluation, error)
}

type LocalSubmissionRepo struct {
	cache cache.LocalSubmitCache
}

func NewLocalSubmitRepo(cache cache.LocalSubmitCache) LocalSubmitRepo {
	return &LocalSubmissionRepo{
		cache: cache,
	}
}

func (repo *LocalSubmissionRepo) StoreEvaluationResult(ctx context.Context, userId, problemId uint64, evals []domain.Evaluation) error {
	return repo.cache.Set(ctx, userId, problemId, evals)
}

func (repo *LocalSubmissionRepo) AcquireEvaluationResult(ctx context.Context, userId, problemId uint64) ([]domain.Evaluation, error) {
	return repo.cache.Get(ctx, userId, problemId)
}
