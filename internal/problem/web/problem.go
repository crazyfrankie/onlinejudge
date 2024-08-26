package web

import "github.com/gin-gonic/gin"

type ProblemHandler struct {
}

func NewProblemHandler() *ProblemHandler {
	return &ProblemHandler{}
}

func (ctl *ProblemHandler) RegisterRoute(r *gin.Engine) {
	problemGroup := r.Group("/problem")
	{
		problemGroup.POST("/create")
	}
}
