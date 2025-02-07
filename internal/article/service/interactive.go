package service

import (
	"context"
	"strconv"

	"github.com/crazyfrankie/onlinejudge/common/constant"
	"github.com/crazyfrankie/onlinejudge/common/errors"
	"github.com/crazyfrankie/onlinejudge/internal/article/domain"
	"github.com/crazyfrankie/onlinejudge/internal/article/repository"
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
		return errors.NewBizError(constant.ErrInteractiveInternalServer)
	}

	return nil
}

func (svc *InteractiveService) Like(ctx context.Context, biz string, bizId, uid uint64) error {
	err := svc.repo.IncrLikeCnt(ctx, biz, bizId, uid)
	if err != nil {
		return errors.NewBizError(constant.ErrInteractiveInternalServer)
	}

	return nil
}

func (svc *InteractiveService) CancelLike(ctx context.Context, biz string, bizId, uid uint64) error {
	err := svc.repo.DecrLikeCnt(ctx, biz, bizId, uid)
	if err != nil {
		return errors.NewBizError(constant.ErrInteractiveInternalServer)
	}

	return nil
}

func (svc *InteractiveService) GetInteractive(ctx context.Context, biz string, bizId string, uid uint64) (domain.Interactive, error) {
	bizID, _ := strconv.Atoi(bizId)
	inter, err := svc.repo.GetInteractive(ctx, biz, uint64(bizID), uid)
	if err != nil {
		return domain.Interactive{}, errors.NewBizError(constant.ErrInteractiveInternalServer)
	}

	return inter, nil
}
