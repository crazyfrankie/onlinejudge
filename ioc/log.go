package ioc

import (
	"go.uber.org/zap"

	"github.com/crazyfrankie/onlinejudge/pkg/zapx"
)

func InitLog() *zapx.Logger {
	l := zapx.NewLogger(zap.NewProductionConfig())
	return l
}
