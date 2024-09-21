package repository

import (
	"context"
	"errors"
	"oj/internal/judgement/repository/cache"
)

type SubmitRepository interface {
	StoreEvaluation(ctx context.Context, result string) error
}

type SubmissionRepo struct {
	cache cache.SubmitCache
}

func NewSubmitRepository(cache cache.SubmitCache) SubmitRepository {
	return &SubmissionRepo{
		cache: cache,
	}
}

func (repo *SubmissionRepo) StoreEvaluation(ctx context.Context, result string) error {
	return errors.New("error")
}
