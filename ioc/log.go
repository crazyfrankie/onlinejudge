package ioc

import "go.uber.org/zap"

func InitLog() *zap.Logger {
	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return log
}
