package service

import (
	"context"
	"oj/common/constant"
	"oj/common/errors"
	"strconv"

	"oj/internal/article/repository"
)

type InteractiveService struct {
	repo *repository.InteractiveArtRepository
}

func NewInteractiveService(repo *repository.InteractiveArtRepository) *InteractiveService {
	return &InteractiveService{
		repo: repo,
	}
}

func (svc *InteractiveService) IncrReadCnt(ctx context.Context, biz, bizId string) error {
	bizID, _ := strconv.Atoi(bizId)
	err := svc.repo.IncrReadCnt(ctx, biz, uint64(bizID))
	if err != nil {
		return errors.NewBusinessError(constant.ErrInternalServer)
	}

	return nil
}

func (svc *InteractiveService) IncrLikeCnt(ctx context.Context, biz, bizId string) error {
	bizID, _ := strconv.Atoi(bizId)
	err := svc.repo.IncrLikeCnt(ctx, biz, uint64(bizID))
	if err != nil {
		return errors.NewBusinessError(constant.ErrInternalServer)
	}

	return nil
}

func (svc *InteractiveService) IncrCollectCnt(ctx context.Context, biz, bizId string) error {
	bizID, _ := strconv.Atoi(bizId)
	err := svc.repo.IncrCollectCnt(ctx, biz, uint64(bizID))
	if err != nil {
		return errors.NewBusinessError(constant.ErrInternalServer)
	}

	return nil
}
