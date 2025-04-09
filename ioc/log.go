package ioc

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/crazyfrankie/onlinejudge/pkg/zapx"
)

func InitLog() *zap.Logger {
	cfg := zap.NewProductionConfig()
	l, err := cfg.Build(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapx.NewCustomCore(core)
	}))
	if err != nil {
		return nil
	}
	zap.RedirectStdLog(l)

	return l
}
