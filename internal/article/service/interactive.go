package service

import (
	"context"
	"strconv"

	"oj/common/constant"
	"oj/common/errors"
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

func (svc *InteractiveService) Like(ctx context.Context, biz string, bizId, uid uint64) error {
	err := svc.repo.IncrLikeCnt(ctx, biz, bizId, uid)
	if err != nil {
		return errors.NewBusinessError(constant.ErrInternalServer)
	}

	return nil
}

func (svc *InteractiveService) CancelLike(ctx context.Context, biz string, bizId, uid uint64) error {
	err := svc.repo.DecrLikeCnt(ctx, biz, bizId, uid)
	if err != nil {
		return errors.NewBusinessError(constant.ErrInternalServer)
	}

	return nil
}
