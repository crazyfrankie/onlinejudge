package repository

import (
	"context"

	"oj/internal/article/repository/dao"
)

type InteractiveArtRepository struct {
	dao *dao.InteractiveDao
}

func NewInteractiveArtRepository(dao *dao.InteractiveDao) *InteractiveArtRepository {
	return &InteractiveArtRepository{
		dao: dao,
	}
}

func (r InteractiveArtRepository) IncrReadCnt(ctx context.Context, biz string, bizId uint64) error {
	return r.dao.IncrReadCnt(ctx, biz, bizId)
}

func (r InteractiveArtRepository) IncrLikeCnt(ctx context.Context, biz string, bizId uint64) error {
	return r.dao.IncrLikeCnt(ctx, biz, bizId)
}

func (r InteractiveArtRepository) IncrCollectCnt(ctx context.Context, biz string, bizId uint64) error {
	return r.dao.IncrLikeCnt(ctx, biz, bizId)
}
