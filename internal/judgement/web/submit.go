package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
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
		submitGroup.POST("evaluate", ctl.GetResult())
	}
}

func (ctl *SubmissionHandler) SubmitCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			UserId    uint64 `json:"userId"`
			ProblemId uint64 `json:"problemId"`
			Code      string `json:"code"`
			Language  string `json:"language"`
		}

		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}

		result, err := ctl.svc.SubmitCode(c.Request.Context(), domain.Submission{
			UserId:    req.UserId,
			ProblemID: req.ProblemId,
			Code:      req.Code,
		}, req.Language)

		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func (ctl *SubmissionHandler) GetResult() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Tokens []string `json:"tokens"`
		}

		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}

		evals, err := ctl.svc.GetResult(c.Request.Context(), req.Tokens)
		if err != nil {
			c.JSON(http.StatusBadRequest, "system error")
			return
		}

		c.JSON(http.StatusOK, evals)
	}
}
