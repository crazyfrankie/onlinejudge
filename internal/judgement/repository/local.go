package repository

import (
	"context"
	"errors"
	"oj/internal/judgement/domain"
	"oj/internal/judgement/repository/cache"
)

type LocalSubmitRepo interface {
	StoreEvaluationResult(ctx context.Context, userId, problemId uint64, evals []domain.Evaluation) error
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
	return errors.New("error")
}
