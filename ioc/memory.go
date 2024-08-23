package ioc

import (
	"github.com/patrickmn/go-cache"
	"time"
)

func InitGoMem() *cache.Cache {
	return cache.New(time.Minute*10, time.Minute*15)
}
