package ioc

import (
	"github.com/crazyfrankie/onlinejudge/config"

	"github.com/crazyfrankie/go-judge/pkg/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitJudgeClient() rpc.JudgeServiceClient {
	cc, err := grpc.NewClient(config.GetConf().Judge.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil
	}

	return rpc.NewJudgeServiceClient(cc)
}
