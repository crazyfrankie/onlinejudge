package ioc

import (
	"oj/internal/user/service/sms"
	"oj/internal/user/service/sms/memory"
)

func InitSMSService() sms.Service {
	return memory.NewService()
}
