package service

import (
	"context"
	"oj/internal/article/domain"
	"oj/internal/article/repository"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (uint64, error)
	Publish(ctx context.Context, art domain.Article) (uint64, error)
}

type articleService struct {
	repo *repository.ArticleRepository
}

func NewArticleService(repo *repository.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
	}
}

func (svc *articleService) Save(ctx context.Context, art domain.Article) (uint64, error) {
	if art.ID > 0 {
		err := svc.repo.Update(ctx, art)
		return art.ID, err
	}
	return svc.repo.Create(ctx, art)
}

func (svc *articleService) Publish(ctx context.Context, art domain.Article) (uint64, error) {
	//TODO implement me
	panic("implement me")
}
