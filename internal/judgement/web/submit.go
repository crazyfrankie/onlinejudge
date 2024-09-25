package web

import (
	"errors"
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
		submitGroup.POST("run", ctl.RunCode())
		submitGroup.POST("submit", ctl.SubmitCode())
	}
}

func (ctl *SubmissionHandler) RunCode() gin.HandlerFunc {
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

		result, err := ctl.svc.RunCode(c.Request.Context(), domain.Submission{
			UserId:    req.UserId,
			ProblemID: req.ProblemId,
			Code:      req.Code,
		}, req.Language)

		switch {
		case errors.Is(err, service.ErrSyntax):
			c.JSON(http.StatusBadRequest, "your code not fit format")
			return
		case err != nil:
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		c.JSON(http.StatusOK, result)
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

		result, err := ctl.svc.RunCode(c.Request.Context(), domain.Submission{
			UserId:    req.UserId,
			ProblemID: req.ProblemId,
			Code:      req.Code,
		}, req.Language)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "system error")
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
