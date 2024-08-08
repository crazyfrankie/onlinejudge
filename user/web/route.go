package web

import (
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/contrib/cors"
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

	router.Use(cors.New(cors.Config{
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		// 是否允许带 cookie 之类的
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "localhost") {
				// 开发环境
				return true
			}
			return strings.Contains(origin, "yourcompany.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	// 采用 memstore 进行数据存储

	//store := memstore.NewStore([]byte("KsS2X1CgFT4bi3BRRIxLk5jjiUBj8wxE"),
	//	[]byte("8nGgE3Uz9EHMAgNr2PxFKCgM2V08SF2h"))

	// 采用 redis 进行数据存储
	store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
		[]byte("KsS2X1CgFT4bi3BRRIxLk5jjiUBj8wxE"),
		[]byte("8nGgE3Uz9EHMAgNr2PxFKCgM2V08SF2h"))
	if err != nil {
		panic(err)
	}

	//myStore := &sqlx_store.Store{}
	//router.Use(sessions.Sessions("sess_name", myStore))

	router.Use(sessions.Sessions("sess_name", store))

	router.Use(middleware.NewLoginMiddlewareBuilder().
		IgnorePaths("/user/signup").
		IgnorePaths("/user/login").
		CheckLogin())

	// 路由注册
	u := InitHandler()
	u.RegisterRoute(router)

	return router
}
