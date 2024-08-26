package user

import (
	"oj/internal/user/service"
	"oj/internal/user/service/sms"
	"oj/internal/user/service/sms/memory"
	"oj/internal/user/web"
)

func initHandler(codeSvc service.CodeService, userSvc service.UserService) *web.UserHandler {
	return web.NewUserHandler(userSvc, codeSvc)
}

func InitSMSService() sms.Service {
	// 到时候修改这里即可 比如 tencent.NewService()
	return memory.NewService()
}
