package service

import (
	"context"
	"errors"

	"github.com/crazyfrankie/onlinejudge/common/constant"
	er "github.com/crazyfrankie/onlinejudge/common/errors"
	"github.com/crazyfrankie/onlinejudge/internal/problem/domain"
	"github.com/crazyfrankie/onlinejudge/internal/problem/repository"
	"strconv"
)

var (
	ErrProblemNotFound = repository.ErrProblemNotFound
	ErrProblemExists   = repository.ErrProblemExists
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
	err := svc.repo.InsertProblem(ctx, problem)
	if err != nil {
		if errors.Is(err, ErrProblemExists) {
			return er.NewBizError(constant.ErrProblemExists)
		}

		return er.NewBizError(constant.ErrProblemInternalServer)
	}

	return nil
}

func (svc *ProblemSvc) ModifyProblem(ctx context.Context, id string, problem domain.Problem) (domain.Problem, error) {
	Id, err := strconv.Atoi(id)
	if err != nil {
		return domain.Problem{}, err
	}

	var pm domain.Problem
	pm, err = svc.repo.UpdateProblem(ctx, uint64(Id), problem)
	if err != nil {
		if errors.Is(err, repository.ErrProblemNotFound) {
			return domain.Problem{}, er.NewBizError(constant.ErrProblemNotFound)
		}

		return domain.Problem{}, er.NewBizError(constant.ErrProblemInternalServer)
	}

	return pm, nil
}

func (svc *ProblemSvc) GetAllProblems(ctx context.Context) ([]domain.Problem, error) {
	problems, err := svc.repo.FindAllProblems(ctx)
	if err != nil {
		return []domain.Problem{}, er.NewBizError(constant.ErrProblemInternalServer)
	}

	return problems, nil
}

func (svc *ProblemSvc) AddTag(ctx context.Context, tag string) error {
	err := svc.repo.CreateTag(ctx, tag)
	if err != nil {
		if errors.Is(err, repository.ErrTagExists) {
			return er.NewBizError(constant.ErrProblemTagExists)
		}

		return er.NewBizError(constant.ErrProblemInternalServer)
	}

	return nil
}

func (svc *ProblemSvc) ModifyTag(ctx context.Context, id uint64, newTag string) error {
	err := svc.repo.UpdateTag(ctx, id, newTag)
	if err != nil {
		if errors.Is(err, ErrTagExists) {
			return er.NewBizError(constant.ErrProblemTagExists)
		}

		return er.NewBizError(constant.ErrProblemInternalServer)
	}

	return nil
}

func (svc *ProblemSvc) FindAllTags(ctx context.Context) ([]domain.Tag, error) {
	tags, err := svc.repo.FindAllTags(ctx)
	if err != nil {
		if errors.Is(err, ErrNoTags) {
			return []domain.Tag{}, er.NewBizError(constant.ErrProblemNoTags)
		}

		return []domain.Tag{}, er.NewBizError(constant.ErrProblemInternalServer)
	}

	return tags, nil
}

func (svc *ProblemSvc) FindCountByTags(ctx context.Context) ([]domain.TagWithCount, error) {
	tagCount, err := svc.repo.FindCountInTag(ctx)
	if err != nil {
		if errors.Is(err, ErrNoTags) {
			return []domain.TagWithCount{}, er.NewBizError(constant.ErrProblemNoTags)
		}

		return []domain.TagWithCount{}, er.NewBizError(constant.ErrProblemInternalServer)
	}

	return tagCount, nil
}

func (svc *ProblemSvc) GetProblemsByTag(ctx context.Context, name string) ([]domain.RoughProblem, error) {
	problems, err := svc.repo.FindProblemsByName(ctx, name)
	if err != nil {
		return []domain.RoughProblem{}, err
	}

	return problems, nil
}

func (svc *ProblemSvc) GetProblem(ctx context.Context, id uint64, tag, title string) (domain.Problem, error) {
	pm, err := svc.repo.FindByTitle(ctx, id, tag, title)
	if err != nil {
		if errors.Is(err, ErrProblemNotFound) {
			return domain.Problem{}, er.NewBizError(constant.ErrProblemNotFound)
		}

		return domain.Problem{}, er.NewBizError(constant.ErrProblemInternalServer)
	}

	return pm, nil
}
