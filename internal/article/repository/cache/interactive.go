package cache

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/redis/go-redis/v9"
)

//go:embed lua/interactive.lua
var incrCntLua string

type InteractiveCache struct {
	cmd redis.Cmdable
}

func NewInteractiveCache(cmd redis.Cmdable) *InteractiveCache {
	return &InteractiveCache{cmd: cmd}
}

func (cache *InteractiveCache) IncrReadCnt(ctx context.Context, biz string, bizId uint64) error {
	return cache.cmd.Eval(ctx, incrCntLua, []string{cache.key(biz, bizId)}, "read_cnt", 1).Err()
}

func (cache *InteractiveCache) IncrLikeCnt(ctx context.Context, biz string, bizId uint64) error {
	return cache.cmd.Eval(ctx, incrCntLua, []string{cache.key(biz, bizId)}, "like_cnt", 1).Err()
}

func (cache *InteractiveCache) DecrLikeCnt(ctx context.Context, biz string, bizId uint64) error {
	return cache.cmd.Eval(ctx, incrCntLua, []string{cache.key(biz, bizId)}, "like_cnt", -1).Err()
}

func (cache *InteractiveCache) key(biz string, bizId uint64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}
