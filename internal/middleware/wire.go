//go:build wireinject

package middleware

import (
	"github.com/crazyfrankie/onlinejudge/internal/middleware/jwt"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

func InitModule(cmd redis.Cmdable) *Module {
	wire.Build(
		jwt.NewRedisJWTHandler,

		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
