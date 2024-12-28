package web

import (
	"errors"
	"net/http"

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
		case errors.Is(err, remote.ErrSyntax):
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("your code not fit format")))
			return

		case err != nil:
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system errors")))
			return
		}

		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithData(result)))
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
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system errors")))
			return
		}

		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithData(result)))
	}
}
