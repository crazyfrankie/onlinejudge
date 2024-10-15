package web

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"oj/internal/problem/domain"
	"oj/internal/problem/service"
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
	modifyGroup := r.Group("admin/problem")
	{
		modifyGroup.POST("create", ctl.AddProblem())
		modifyGroup.GET("")
		modifyGroup.PUT("modify/:id", ctl.ModifyProblem())
	}

	// 题目获取
	getGroup := r.Group("")
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
		type TestCaseReq struct {
			Input  string `json:"input"`
			Output string `json:"output"`
		}

		type Req struct {
			UserId     uint64        `json:"userId"`
			Title      string        `json:"title"`
			Tag        string        `json:"tag"`
			Content    string        `json:"content"`
			Prompt     []string      `json:"prompt"`
			TestCases  []TestCaseReq `json:"testCases"`
			PassRate   string        `json:"passRate"`
			MaxMem     int           `json:"maxMem"`
			MaxRunTime int           `json:"maxRunTime"`
			Difficulty uint8         `json:"difficulty"`
		}

		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}

		domainTestCases := make([]domain.TestCase, len(req.TestCases))
		for i, tc := range req.TestCases {
			domainTestCases[i] = domain.TestCase{
				Input:  tc.Input,
				Output: tc.Output,
			}
		}

		pm := domain.Problem{
			UserId:     req.UserId,
			Title:      req.Title,
			Tag:        req.Tag,
			Content:    req.Content,
			Prompt:     make([]string, len(req.Prompt)),
			TestCases:  domainTestCases,
			PassRate:   req.PassRate,
			MaxMem:     req.MaxMem,
			MaxRuntime: req.MaxRunTime,
			Difficulty: req.Difficulty,
		}
		copy(pm.Prompt, req.Prompt)

		err := ctl.svc.AddProblem(c.Request.Context(), pm)
		switch {
		case err != nil:
			c.JSON(http.StatusBadRequest, "system error")
		default:
			c.JSON(http.StatusOK, "add successfully")
		}
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

		switch {
		case errors.Is(err, service.ErrProblemNotFound):
			c.JSON(http.StatusBadRequest, "problem not found")
		case err != nil:
			c.JSON(http.StatusInternalServerError, "system error")
		default:
			c.JSON(http.StatusOK, pm)
		}
	}
}

func (ctl *ProblemHandler) GetAllProblems() gin.HandlerFunc {
	return func(c *gin.Context) {
		problems, err := ctl.svc.GetAllProblems(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, "system error")
			return
		}

		c.JSON(http.StatusOK, problems)
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

		switch {
		case errors.Is(err, service.ErrTagExists):
			c.JSON(http.StatusBadRequest, "tag already exists")
		case err != nil:
			c.JSON(http.StatusInternalServerError, "system error")
		default:
			c.JSON(http.StatusOK, "add successfully")
		}
	}
}

func (ctl *ProblemHandler) GetAllTags() gin.HandlerFunc {
	return func(c *gin.Context) {
		tags, err := ctl.svc.FindAllTags(c.Request.Context())

		switch {
		case errors.Is(err, service.ErrNoTags):
			c.JSON(http.StatusBadRequest, "tag need to be created")
		case err != nil:
			c.JSON(http.StatusInternalServerError, "system error")
		default:
			c.JSON(http.StatusOK, tags)
		}
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
		switch {
		case errors.Is(err, service.ErrTagExists):
			c.JSON(http.StatusBadRequest, "this tag already exists")
		case err != nil:
			c.JSON(http.StatusInternalServerError, "system error")
		default:
			c.JSON(http.StatusOK, req.Tag)
		}
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
		switch {
		case errors.Is(err, service.ErrProblemNotFound):
			c.JSON(http.StatusNotFound, "problem not found")
		case err != nil:
			c.JSON(http.StatusInternalServerError, "system error")
		default:
			c.JSON(http.StatusOK, pm)
		}
	}
}

func (ctl *ProblemHandler) GetPmListByCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		tagName := c.Param("tag")

		problems, err := ctl.svc.GetProblemsByTag(c.Request.Context(), tagName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "system error")
		}

		c.JSON(http.StatusOK, problems)
	}
}

func (ctl *ProblemHandler) GetProblemSet() gin.HandlerFunc {
	return func(c *gin.Context) {
		tags, err := ctl.svc.FindCountByTags(c.Request.Context())

		switch {
		case errors.Is(err, service.ErrNoTags):
			c.JSON(http.StatusBadRequest, "tag need to be created")
		case err != nil:
			c.JSON(http.StatusInternalServerError, "system error")
		default:
			c.JSON(http.StatusOK, tags)
		}
	}
}
