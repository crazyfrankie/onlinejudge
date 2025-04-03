package web

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/crazyfrankie/onlinejudge/common/response"
	"github.com/crazyfrankie/onlinejudge/internal/problem/domain"
	"github.com/crazyfrankie/onlinejudge/internal/problem/service"
)

type ProblemHandler struct {
	svc service.ProblemService
}

func NewProblemHandler(svc service.ProblemService) *ProblemHandler {
	return &ProblemHandler{
		svc: svc,
	}
}

func (ctl *ProblemHandler) RegisterRoute(r *gin.Engine) {
	//  管理员题目的增改查
	modifyGroup := r.Group("api/admin/problem")
	{
		modifyGroup.POST("create", ctl.AddProblem())
		modifyGroup.GET("")
		modifyGroup.PUT("modify/:id", ctl.ModifyProblem())
	}

	// 题目获取
	getGroup := r.Group("api/")
	{
		getGroup.GET("problemset", ctl.GetProblemSet())              // 获取所有分类问题集
		getGroup.GET("problem-list/:tag", ctl.GetPmListByCategory()) // 获取特定分类的问题集
		getGroup.GET("problems/:name/description", ctl.GetProblem()) // 获取某个问题的详细信息
	}

	// 标签的增查改
	tagGroup := r.Group("tags")
	{
		tagGroup.POST("add", ctl.AddTag())
		tagGroup.GET("", ctl.GetAllTags())
		tagGroup.PUT("modify", ctl.ModifyTag())
	}

}

func (ctl *ProblemHandler) AddProblem() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			UserId     uint64                 `json:"user_id"`
			Title      string                 `json:"title"`
			Tag        string                 `json:"tag"`
			Content    string                 `json:"content"`
			TestCases  []domain.LocalTestCase `json:"test_case"`
			PassRate   string                 `json:"pass_rate"`
			FuncName   string                 `json:"func_name"`
			Params     string                 `json:"params"`
			PreDefine  string                 `json:"pre_define"`
			MaxMem     int                    `json:"max_mem"`
			MaxRunTime int                    `json:"max_run_time"`
			Difficulty uint8                  `json:"difficulty"`
		}

		var req Req
		if err := c.Bind(&req); err != nil {
			zap.L().Error("添加题目:绑定失败", zap.Error(err))
			return
		}

		pm := domain.Problem{
			UserId:     req.UserId,
			Title:      req.Title,
			Tag:        req.Tag,
			Content:    req.Content,
			PassRate:   req.PassRate,
			TestCases:  req.TestCases,
			MaxMem:     req.MaxMem,
			FuncName:   req.FuncName,
			Params:     req.Params,
			PreDefine:  req.PreDefine,
			MaxRuntime: req.MaxRunTime,
			Difficulty: req.Difficulty,
		}

		err := ctl.svc.AddProblem(c.Request.Context(), pm)
		if err != nil {
			response.Error(c, err)
		}

		response.Success(c, nil)
	}
}

func (ctl *ProblemHandler) ModifyProblem() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Title      string `json:"title"`
			Content    string `json:"content"`
			Difficulty uint8  `json:"difficulty"`
		}

		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}

		id := c.Param("id")

		pm, err := ctl.svc.ModifyProblem(c.Request.Context(), id, domain.Problem{
			Title:      req.Title,
			Content:    req.Content,
			Difficulty: req.Difficulty,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, pm)

	}
}

func (ctl *ProblemHandler) GetAllProblems() gin.HandlerFunc {
	return func(c *gin.Context) {
		problems, err := ctl.svc.GetAllProblems(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, problems)
	}
}

func (ctl *ProblemHandler) AddTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Tag string `json:"tag"`
		}
		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}

		err := ctl.svc.AddTag(c.Request.Context(), req.Tag)

		if err != nil {
			response.Error(c, err)
		}

		response.Success(c, nil)
	}
}

func (ctl *ProblemHandler) GetAllTags() gin.HandlerFunc {
	return func(c *gin.Context) {
		tags, err := ctl.svc.FindAllTags(c.Request.Context())

		if err != nil {
			response.Error(c, err)
		}

		response.Success(c, tags)
	}
}

func (ctl *ProblemHandler) ModifyTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Id  uint64 `json:"id"`
			Tag string `json:"tag"`
		}

		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}

		err := ctl.svc.ModifyTag(c.Request.Context(), req.Id, req.Tag)
		if err != nil {
			response.Error(c, err)
		}

		response.Success(c, nil)
	}
}

func (ctl *ProblemHandler) GetProblem() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Id  uint64 `json:"id"`
			Tag string `json:"tag"`
		}
		var req Req
		title := c.Param("name")

		if err := c.Bind(&req); err != nil {
			return
		}

		pm, err := ctl.svc.GetProblem(c.Request.Context(), req.Id, req.Tag, title)
		if err != nil {
			response.Error(c, err)
		}

		response.Success(c, pm)
	}
}

func (ctl *ProblemHandler) GetPmListByCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		tagName := c.Param("tag")

		problems, err := ctl.svc.GetProblemsByTag(c.Request.Context(), tagName)
		if err != nil {
			response.Error(c, err)
		}

		response.Success(c, problems)
	}
}

func (ctl *ProblemHandler) GetProblemSet() gin.HandlerFunc {
	return func(c *gin.Context) {
		tags, err := ctl.svc.FindCountByTags(c.Request.Context())
		if err != nil {
			response.Error(c, err)
		}

		response.Success(c, tags)
	}
}
