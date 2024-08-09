package web

import (
	"github.com/gin-gonic/gin"

	"oj/user/repository"
	"oj/user/repository/dao"
	"oj/user/service"
	"oj/user/web/middleware"
)

// InitHandler Handler 对象的创建
func InitHandler() *UserHandler {
	db := dao.InitDB()
	db.AutoMigrate(&dao.User{})

	ud := dao.NewUserDao(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := NewUserHandler(svc)
	return u
}

// InitWeb gin 框架的初始化以及路由的注册
func InitWeb() *gin.Engine {
	router := gin.Default()

	// 跨域设置
	router.Use(middleware.Cors())

	// 登录校验
	router.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/user/signup").
		IgnorePaths("/user/login").
		CheckLogin())

	// 路由注册
	u := InitHandler()
	u.RegisterRoute(router)

	return router
}
