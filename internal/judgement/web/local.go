package web

import (
	"github.com/crazyfrankie/onlinejudge/internal/judgement/service/local"
	"github.com/gin-gonic/gin"
)

type LocalSubmitHandler struct {
	svc local.LocSubmitService
}

func NewLocalSubmitHandler(svc local.LocSubmitService) *LocalSubmitHandler {
	return &LocalSubmitHandler{
		svc: svc,
	}
}

func (ctl *LocalSubmitHandler) RegisterRoute(r *gin.Engine) {
	//submitGroup := r.Group("api/local")
	{
		//submitGroup.POST("run", ctl.RunCode())
		//submitGroup.POST("submit", ctl.SubmitCode())
	}
}

//func (ctl *LocalSubmitHandler) RunCode() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		type Req struct {
//			UserId    uint64 `json:"userId"`
//			ProblemId uint64 `json:"problemId"`
//			Code      string `json:"code"`
//			Language  string `json:"language"`
//		}
//		var req Req
//		if err := c.Bind(&req); err != nil {
//			return
//		}
//
//		results, err := ctl.svc.RunCode(c.Request.Context(), domain.Submission{
//			UserId:    req.UserId,
//			ProblemID: req.ProblemId,
//			Code:      req.Code,
//		}, req.Language)
//
//		switch {
//		case errors.Is(err, remote.ErrSyntax):
//			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("your code not fit format")))
//			return
//
//		case err != nil:
//			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system errors")))
//			return
//		}
//
//		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithData(results)))
//	}
//}
//
//func (ctl *LocalSubmitHandler) SubmitCode() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		type Req struct {
//			UserId    uint64 `json:"userId"`
//			ProblemId uint64 `json:"problemId"`
//			Code      string `json:"code"`
//			Language  string `json:"language"`
//		}
//
//		var req Req
//		if err := c.Bind(&req); err != nil {
//			return
//		}
//
//		result, err := ctl.svc.RunCode(c.Request.Context(), domain.Submission{
//			UserId:    req.UserId,
//			ProblemID: req.ProblemId,
//			Code:      req.Code,
//		}, req.Language)
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system errors")))
//			return
//		}
//
//		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithData(result)))
//	}
//}
