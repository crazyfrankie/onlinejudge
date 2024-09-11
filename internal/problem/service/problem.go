package service

import (
	"context"
	"oj/internal/problem/domain"
	"oj/internal/problem/repository"
	"strconv"
)

var (
	ErrProblemNotFound = repository.ErrProblemNotFound
	ErrTagExists       = repository.ErrTagExists
	ErrNoTags          = repository.ErrNoTags
)

type ProblemService interface {
	AddProblem(ctx context.Context, problem domain.Problem) error
	ModifyProblem(ctx context.Context, id string, problem domain.Problem) (domain.Problem, error)
	GetAllProblems(ctx context.Context) ([]domain.Problem, error)
	AddTag(ctx context.Context, tag string) error
	ModifyTag(ctx context.Context, id uint64, newTag string) error
	FindCountByTags(ctx context.Context) ([]domain.TagWithCount, error)
	FindAllTags(ctx context.Context) ([]domain.Tag, error)
	GetProblemsByTag(ctx context.Context, name string) ([]domain.RoughProblem, error)
	GetProblem(ctx context.Context, id uint64, tag, title string) (domain.Problem, error)
}

type ProblemSvc struct {
	repo repository.ProblemRepository
}

func NewProblemService(repo repository.ProblemRepository) ProblemService {
	return &ProblemSvc{
		repo: repo,
	}
}

func (svc *ProblemSvc) AddProblem(ctx context.Context, problem domain.Problem) error {
	return svc.repo.InsertProblem(ctx, problem)
}

func (svc *ProblemSvc) ModifyProblem(ctx context.Context, id string, problem domain.Problem) (domain.Problem, error) {
	Id, err := strconv.Atoi(id)
	if err != nil {
		return domain.Problem{}, err
	}

	return svc.repo.UpdateProblem(ctx, uint64(Id), problem)
}

func (svc *ProblemSvc) GetAllProblems(ctx context.Context) ([]domain.Problem, error) {
	return svc.repo.FindAllProblems(ctx)
}

func (svc *ProblemSvc) AddTag(ctx context.Context, tag string) error {
	return svc.repo.CreateTag(ctx, tag)
}

func (svc *ProblemSvc) ModifyTag(ctx context.Context, id uint64, newTag string) error {
	return svc.repo.UpdateTag(ctx, id, newTag)
}

func (svc *ProblemSvc) FindAllTags(ctx context.Context) ([]domain.Tag, error) {
	return svc.repo.FindAllTags(ctx)
}

func (svc *ProblemSvc) FindCountByTags(ctx context.Context) ([]domain.TagWithCount, error) {
	return svc.repo.FindCountInTag(ctx)
}

func (svc *ProblemSvc) GetProblemsByTag(ctx context.Context, name string) ([]domain.RoughProblem, error) {
	return svc.repo.FindProblemsByName(ctx, name)
}

func (svc *ProblemSvc) GetProblem(ctx context.Context, id uint64, tag, title string) (domain.Problem, error) {
	return svc.repo.FindByTitle(ctx, id, tag, title)
}
