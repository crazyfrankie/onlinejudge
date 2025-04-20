package web

import (
	"github.com/crazyfrankie/onlinejudge/common/response"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/onlinejudge/internal/judgement/domain"
	"github.com/crazyfrankie/onlinejudge/internal/judgement/service/remote"
)

type SubmissionHandler struct {
	svc remote.SubmitService
}

func NewSubmissionHandler(svc remote.SubmitService) *SubmissionHandler {
	return &SubmissionHandler{
		svc: svc,
	}
}

func (ctl *SubmissionHandler) RegisterRoute(r *gin.Engine) {
	submitGroup := r.Group("api/remote")
	{
		submitGroup.POST("run", ctl.RunCode())
		submitGroup.POST("submit", ctl.SubmitCode())
	}
}

func (ctl *SubmissionHandler) RunCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := "onlinejudge/Judge/Remote/RunCode"
		type Req struct {
			UserId    uint64 `json:"userId"`
			ProblemId uint64 `json:"problemId"`
			Code      string `json:"code"`
			Language  string `json:"language"`
		}

		var req Req
		if err := c.Bind(&req); err != nil {
			response.ErrorWithLog(c, name, "bind req error", err)
			return
		}

		result, err := ctl.svc.RunCode(c.Request.Context(), domain.Submission{
			UserId:    req.UserId,
			ProblemID: req.ProblemId,
			Code:      req.Code,
		}, req.Language)
		if err != nil {
			response.ErrorWithLog(c, name, bizError, err)
			return
		}

		response.SuccessWithLog(c, result, name, success)
	}
}

func (ctl *SubmissionHandler) SubmitCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := "onlinejudge/Judge/Remote/SubmitCode"

		type Req struct {
			UserId    uint64 `json:"userId"`
			ProblemId uint64 `json:"problemId"`
			Code      string `json:"code"`
			Language  string `json:"language"`
		}

		var req Req
		if err := c.Bind(&req); err != nil {
			response.ErrorWithLog(c, name, "bind req error", err)
			return
		}

		result, err := ctl.svc.RunCode(c.Request.Context(), domain.Submission{
			UserId:    req.UserId,
			ProblemID: req.ProblemId,
			Code:      req.Code,
		}, req.Language)
		if err != nil {
			response.ErrorWithLog(c, name, bizError, err)
			return
		}

		response.SuccessWithLog(c, result, name, success)
	}
}
