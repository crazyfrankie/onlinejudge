package repository

import (
	"context"

	"github.com/crazyfrankie/onlinejudge/internal/judgement/domain"
	"github.com/crazyfrankie/onlinejudge/internal/judgement/repository/cache"
	"github.com/crazyfrankie/onlinejudge/internal/judgement/repository/dao"
)

type LocalSubmitRepo interface {
	CreateSubmit(ctx context.Context, sub domain.Submission) (uint64, error)
	CreateEvaluate(ctx context.Context, eva domain.Evaluation) error
	UpdateEvaluate(ctx context.Context, pid, sid uint64, state string) error
	FindEvaluate(ctx context.Context, sid uint64) (domain.Evaluation, error)
}

type LocalSubmissionRepo struct {
	dao   *dao.SubmitDao
	cache cache.LocalSubmitCache
}

func NewLocalSubmitRepo(cache cache.LocalSubmitCache, dao *dao.SubmitDao) LocalSubmitRepo {
	return &LocalSubmissionRepo{
		cache: cache,
		dao:   dao,
	}
}

func (r *LocalSubmissionRepo) CreateSubmit(ctx context.Context, sub domain.Submission) (uint64, error) {
	return r.dao.CreateSubmit(ctx, sub)
}

func (r *LocalSubmissionRepo) CreateEvaluate(ctx context.Context, eva domain.Evaluation) error {
	return r.dao.CreateEvaluate(ctx, eva)
}

func (r *LocalSubmissionRepo) UpdateEvaluate(ctx context.Context, pid, sid uint64, state string) error {
	return r.dao.UpdateEvaluate(ctx, pid, sid, state)
}

func (r *LocalSubmissionRepo) FindEvaluate(ctx context.Context, sid uint64) (domain.Evaluation, error) {
	return r.dao.FindEvaluate(ctx, sid)
}
