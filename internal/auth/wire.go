//go:build wireinject

package auth

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

type Module struct {
	Hdl JWTHandler
}

func InitModule(cmd redis.Cmdable) *Module {
	wire.Build(
		NewRedisJWTHandler,

		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
