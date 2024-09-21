package ioc

import (
	"oj/internal/user/service/oauth/wechat"
	"os"
)

func InitWechatService() wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("environment variable appId not found")
	}

	var appKey string
	appKey, ok = os.LookupEnv("WECHAT_APP_KEY")
	if !ok {
		panic("environment variable appKey not found")
	}
	return wechat.NewService(appId, appKey)
}
