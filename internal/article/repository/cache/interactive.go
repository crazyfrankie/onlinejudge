package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/crazyfrankie/onlinejudge/internal/article/domain"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

//go:embed lua/interactive.lua
var incrCntLua string

var (
	ErrKeyNotExists = errors.New("key not found")
)

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

//func (cache *InteractiveCache) DelReadCnt(ctx context.Context, biz string, bizId uint64) error {
//	return cache.cmd.Del(ctx, cache.key(biz, bizId)).Err()
//}

func (cache *InteractiveCache) GetInteractive(ctx context.Context, biz string, bizId uint64) (domain.Interactive, error) {
	data, err := cache.cmd.HGetAll(ctx, cache.key(biz, bizId)).Result()
	if err != nil {
		return domain.Interactive{}, err
	}

	if len(data) == 0 {
		return domain.Interactive{}, ErrKeyNotExists
	}

	likeCnt, _ := strconv.ParseInt(data["like_cnt"], 10, 64)
	readCnt, _ := strconv.ParseInt(data["read_cnt"], 10, 64)
	liked, _ := strconv.ParseBool(data["liked"])

	return domain.Interactive{
		Liked:   liked,
		LikeCnt: likeCnt,
		ReadCnt: readCnt,
	}, nil
}

func (cache *InteractiveCache) SetInteractive(ctx context.Context, biz string, bizId uint64, inter domain.Interactive) error {
	val := []interface{}{
		"read_cnt", inter.ReadCnt,
		"like_cnt", inter.LikeCnt,
		"liked", inter.Liked,
	}

	err := cache.cmd.HSet(ctx, cache.key(biz, bizId), val...).Err()
	if err != nil {
		return err
	}

	// 设置过期时间
	err = cache.cmd.Expire(ctx, cache.key(biz, bizId), time.Minute*5).Err()
	if err != nil {
		return err
	}

	return nil
}

func (cache *InteractiveCache) key(biz string, bizId uint64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}
