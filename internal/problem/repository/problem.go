package repository

import (
	"context"
	"github.com/crazyfrankie/onlinejudge/internal/problem/domain"
	"github.com/crazyfrankie/onlinejudge/internal/problem/repository/cache"
	"log"

	"github.com/crazyfrankie/onlinejudge/internal/problem/repository/dao"
)

var (
	ErrProblemNotFound = dao.ErrProblemNotFound
	ErrTagExists       = dao.ErrTagExists
	ErrProblemExists   = dao.ErrProblemExists
	ErrNoTags          = dao.ErrNoTags
)

type ProblemRepository interface {
	InsertProblem(ctx context.Context, pm domain.Problem) error
	UpdateProblem(ctx context.Context, id uint64, problem domain.Problem) (domain.Problem, error)
	FindAllProblems(ctx context.Context) ([]domain.Problem, error)
	CreateTag(ctx context.Context, tag string) error
	UpdateTag(ctx context.Context, id uint64, newTag string) error
	FindAllTags(ctx context.Context) ([]domain.Tag, error)

	FindCountInTag(ctx context.Context) ([]domain.TagWithCount, error)
	FindProblemsByName(ctx context.Context, name string) ([]domain.RoughProblem, error)
	FindByTitle(ctx context.Context, id uint64, tag, title string) (domain.Problem, error)
	FindProblemByID(ctx context.Context, id uint64) (domain.Problem, error)
	FindTestById(ctx context.Context, id uint64) (domain.TestCase, error)
}

type CacheProblemRepo struct {
	dao   dao.ProblemDao
	cache cache.ProblemCache
}

func NewProblemRepository(dao dao.ProblemDao, cache cache.ProblemCache) ProblemRepository {
	return &CacheProblemRepo{
		dao:   dao,
		cache: cache,
	}
}

func (repo *CacheProblemRepo) InsertProblem(ctx context.Context, pm domain.Problem) error {
	return repo.dao.CreateProblem(ctx, pm)
}

func (repo *CacheProblemRepo) UpdateProblem(ctx context.Context, id uint64, problem domain.Problem) (domain.Problem, error) {
	pm, err := repo.dao.UpdateProblem(ctx, id, problem)
	if err != nil {
		return pm, err
	}

	// 更新缓存
	err = repo.cache.Set(ctx, pm)
	if err != nil {
		log.Printf("failed to update cache for user %d: %v", problem.Id, err)
	}

	return pm, err
}

func (repo *CacheProblemRepo) FindAllProblems(ctx context.Context) ([]domain.Problem, error) {
	return repo.dao.FindAllProblems(ctx)
}

func (repo *CacheProblemRepo) CreateTag(ctx context.Context, tag string) error {
	return repo.dao.CreateTag(ctx, tag)
}

func (repo *CacheProblemRepo) UpdateTag(ctx context.Context, id uint64, newTag string) error {
	return repo.dao.UpdateTag(ctx, id, newTag)
}

func (repo *CacheProblemRepo) FindAllTags(ctx context.Context) ([]domain.Tag, error) {
	return repo.dao.FindAllTags(ctx)
}

func (repo *CacheProblemRepo) FindCountInTag(ctx context.Context) ([]domain.TagWithCount, error) {
	return repo.dao.FindCountInTag(ctx)
}

func (repo *CacheProblemRepo) FindProblemsByName(ctx context.Context, name string) ([]domain.RoughProblem, error) {
	return repo.dao.FindProblemsByName(ctx, name)
}

func (repo *CacheProblemRepo) FindByTitle(ctx context.Context, id uint64, tag, title string) (domain.Problem, error) {
	// 先去缓存中找
	pm, err := repo.cache.Get(ctx, id)
	if err == nil {
		return pm, nil
	}

	// 去数据库中找
	pm, err = repo.dao.FindByTitle(ctx, tag, title)
	if err != nil {
		return domain.Problem{}, err
	}

	// 异步更新缓存
	go func() {
		newCtx := context.Background()
		err = repo.cache.Set(newCtx, pm)
		if err != nil {
			log.Fatal(err)
		}
	}()

	return pm, nil
}

func (repo *CacheProblemRepo) FindProblemByID(ctx context.Context, pid uint64) (domain.Problem, error) {
	return repo.dao.FindProblemByID(ctx, pid)
}

func (repo *CacheProblemRepo) FindTestById(ctx context.Context, id uint64) (domain.TestCase, error) {
	return repo.dao.FindTestById(ctx, id)
}
