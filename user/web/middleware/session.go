package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

func SessStore() gin.HandlerFunc {
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
	return sessions.Sessions("sess_name", store)
}
