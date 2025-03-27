package web

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/onlinejudge/common/response"
	"github.com/crazyfrankie/onlinejudge/internal/judgement/domain"
	"github.com/crazyfrankie/onlinejudge/internal/judgement/service/local"
	"github.com/crazyfrankie/onlinejudge/internal/middleware/jwt"
)

type SubmitResp struct {
	SubmissionId uint64 `json:"submission_id"`
}

type LocalSubmitHandler struct {
	svc local.LocSubmitService
}

func NewLocalSubmitHandler(svc local.LocSubmitService) *LocalSubmitHandler {
	return &LocalSubmitHandler{
		svc: svc,
	}
}

func (ctl *LocalSubmitHandler) RegisterRoute(r *gin.Engine) {
	submitGroup := r.Group("api/local")
	{
		submitGroup.POST("submit", ctl.RunCode())
		submitGroup.GET("check/:submissionId", ctl.Check())
	}
}

func (ctl *LocalSubmitHandler) RunCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			ProblemId uint64 `json:"problem_id"`
			TypedCode string `json:"typed_code"`
		}
		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}

		claims := c.MustGet("claims")
		claim, _ := claims.(*jwt.Claims)

		submitId, err := ctl.svc.RunCode(c.Request.Context(), domain.Submission{
			ProblemID:  req.ProblemId,
			UserId:     claim.Id,
			Code:       req.TypedCode,
			Language:   "golang",
			SubmitTime: time.Now().Unix(),
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, SubmitResp{
			SubmissionId: submitId,
		})
	}
}

func (ctl *LocalSubmitHandler) Check() gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("submissionId")
		id, _ := strconv.ParseUint(sid, 10, 64)

		res, err := ctl.svc.CheckResult(c.Request.Context(), id)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, res)
	}
}
