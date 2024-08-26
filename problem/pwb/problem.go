package pwb

import "github.com/gin-gonic/gin"

type ProblemHandler struct {
}

func (ctl *ProblemHandler) RegisterRoute(r *gin.Engine) {
	problemGroup := r.Group("/problem")
	{
		problemGroup.POST("/create")
	}
}
