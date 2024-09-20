package web

import (
	"github.com/gin-gonic/gin"
	"oj/internal/judgement/domain"
	"oj/internal/judgement/service"
)

type SubmissionHandler struct {
	svc service.SubmitService
}

func NewSubmissionHandler(svc service.SubmitService) *SubmissionHandler {
	return &SubmissionHandler{
		svc: svc,
	}
}

func (ctl *SubmissionHandler) RegisterRoute(r *gin.Engine) {
	submitGroup := r.Group("")
	{
		submitGroup.POST("submit", ctl.SubmitCode())
	}
}

func (ctl *SubmissionHandler) SubmitCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			UserId    uint64 `json:"userId"`
			ProblemId uint64 `json:"problemId"`
			Code      string `json:"code"`
		}

		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}

		eval, err := ctl.svc.SubmitCode(c.Request.Context(), domain.Submission{
			UserId:    req.UserId,
			ProblemID: req.ProblemId,
			Code:      req.Code,
		})

		if err != nil {

		}
	}
}
