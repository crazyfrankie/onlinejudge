package ioc

import (
	"oj/user/service/sms"
	"oj/user/service/sms/memory"
)

func InitSMSService() sms.Service {
	// 到时候修改这里即可 比如 tencent.NewService()
	return memory.NewService()
}
