package web

import (
	"context"
	"errors"
	"net/http"
	"oj/internal/judgement/domain"
	"oj/internal/judgement/service/remote"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"oj/internal/judgement/service/local"
)

type LocalSubmitHandler struct {
	svc local.LocSubmitService
}

func NewLocalSubmitHandler(svc local.LocSubmitService) *LocalSubmitHandler {
	return &LocalSubmitHandler{
		svc: svc,
	}
}

func (ctl *LocalSubmitHandler) RegisterRoute(r *server.Hertz) {
	submitGroup := r.Group("/local")
	{
		submitGroup.POST("run", ctl.RunCode())
		submitGroup.POST("submit", ctl.SubmitCode())
	}
}

func (ctl *LocalSubmitHandler) RunCode() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
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

		results, err := ctl.svc.RunCode(ctx, domain.Submission{
			UserId:    req.UserId,
			ProblemID: req.ProblemId,
			Code:      req.Code,
		}, req.Language)

		switch {
		case errors.Is(err, remote.ErrSyntax):
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("your code not fit format")))
			return

		case err != nil:
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		}

		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithData(results)))
	}
}

func (ctl *LocalSubmitHandler) SubmitCode() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
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

		result, err := ctl.svc.RunCode(ctx, domain.Submission{
			UserId:    req.UserId,
			ProblemID: req.ProblemId,
			Code:      req.Code,
		}, req.Language)
		if err != nil {
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		}

		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithData(result)))
	}
}
